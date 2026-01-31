package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// API represents the HTTP API server
type API struct {
	store          *Store
	assigner       *TaskAssigner
	workerPool     *AssignmentWorkerPool
	workerPoolCtx  context.Context
	workerPoolStop context.CancelFunc
}

// NewAPI creates a new API instance
func NewAPI() *API {
	store := NewStore()
	assigner := NewTaskAssigner(store)

	// Create worker pool with 5 workers and 30 second timeout
	workerPool := NewAssignmentWorkerPool(assigner, 5, 30*time.Second)

	ctx, cancel := context.WithCancel(context.Background())

	return &API{
		store:          store,
		assigner:       assigner,
		workerPool:     workerPool,
		workerPoolCtx:  ctx,
		workerPoolStop: cancel,
	}
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// CreateEmployeeRequest represents the request body for creating an employee
type CreateEmployeeRequest struct {
	Name     string   `json:"name" binding:"required"`
	Location Location `json:"location" binding:"required"`
	Skills   []string `json:"skills" binding:"required"`
}

// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	Location      Location `json:"location" binding:"required"`
	RequiredSkill string   `json:"required_skill" binding:"required"`
}

// handleCreateEmployee handles POST /employees
func (api *API) handleCreateEmployee(c *gin.Context) {
	var req CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Generate unique ID for the employee
	employee := &Employee{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Location:    req.Location,
		Skills:      req.Skills,
		IsAvailable: true,
	}

	// Validate employee data
	if err := employee.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}

	if err := api.store.AddEmployee(employee); err != nil {
		if taskErr, ok := err.(*TaskError); ok {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   taskErr.Error(),
				Code:    taskErr.Code,
				Message: taskErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Message: "Employee created successfully",
		Data:    employee,
	})
}

// handleCreateTask handles POST /tasks
func (api *API) handleCreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Generate unique ID for the task
	task := &Task{
		ID:            uuid.New().String(),
		Location:      req.Location,
		RequiredSkill: req.RequiredSkill,
		Status:        TaskStatusPending,
	}

	// Validate task data
	if err := task.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}

	// CRITICAL: Submit to queue FIRST to check capacity
	// This prevents orphaned tasks in store if queue is full
	if err := api.workerPool.SubmitTask(task); err != nil {
		// Queue is full, reject request immediately
		if taskErr, ok := err.(*TaskError); ok && taskErr.Code == "QUEUE_FULL" {
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{
				Error:   "System at capacity",
				Code:    "QUEUE_FULL",
				Message: "Worker pool is full. Please retry in a few seconds.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Only add to store AFTER successful queue submission
	if err := api.store.AddTask(task); err != nil {
		if taskErr, ok := err.(*TaskError); ok {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   taskErr.Error(),
				Code:    taskErr.Code,
				Message: taskErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Message: "Task created and assignment initiated",
		Data:    task,
	})
}

// handleGetTasks handles GET /tasks
func (api *API) handleGetTasks(c *gin.Context) {
	tasks := api.store.GetAllTasks()

	c.JSON(http.StatusOK, SuccessResponse{
		Message: fmt.Sprintf("Retrieved %d tasks", len(tasks)),
		Data:    tasks,
	})
}

// handleGetTaskByID handles GET /tasks/:id
func (api *API) handleGetTaskByID(c *gin.Context) {
	taskID := c.Param("id")

	task, err := api.store.GetTask(taskID)
	if err != nil {
		if taskErr, ok := err.(*TaskError); ok {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   taskErr.Error(),
				Code:    taskErr.Code,
				Message: taskErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Task retrieved successfully",
		Data:    task,
	})
}

// handleGetEmployees handles GET /employees
func (api *API) handleGetEmployees(c *gin.Context) {
	api.store.mu.RLock()
	employees := make([]*Employee, 0, len(api.store.employees))
	for _, emp := range api.store.employees {
		employees = append(employees, emp)
	}
	api.store.mu.RUnlock()

	c.JSON(http.StatusOK, SuccessResponse{
		Message: fmt.Sprintf("Retrieved %d employees", len(employees)),
		Data:    employees,
	})
}

// handleHealthCheck handles GET /health
func (api *API) handleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().UTC(),
	})
}

// setupRouter configures all routes
func (api *API) setupRouter() *gin.Engine {
	router := gin.Default()

	// Add CORS middleware for development
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Health check endpoint
	router.GET("/health", api.handleHealthCheck)

	// Employee endpoints
	router.POST("/employees", api.handleCreateEmployee)
	router.GET("/employees", api.handleGetEmployees)

	// Task endpoints
	router.POST("/tasks", api.handleCreateTask)
	router.GET("/tasks", api.handleGetTasks)
	router.GET("/tasks/:id", api.handleGetTaskByID)

	return router
}

// Start starts the API server and worker pool
func (api *API) Start(port string) error {
	// Start worker pool
	api.workerPool.Start(api.workerPoolCtx)
	log.Println("Worker pool started with 5 workers")

	// Setup router
	router := api.setupRouter()

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Stop accepting new tasks
	api.workerPoolStop()

	// Shutdown worker pool and wait for tasks to complete
	log.Println("Waiting for worker pool to drain...")
	api.workerPool.Shutdown()
	log.Println("Worker pool shutdown complete")

	// Shutdown HTTP server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	log.Println("Server exited gracefully")
	return nil
}

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create and start API
	api := NewAPI()
	if err := api.Start(port); err != nil {
		log.Fatal(err)
	}
}
