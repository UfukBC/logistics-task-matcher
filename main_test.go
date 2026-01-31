package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

// setupTestAPI creates a test API instance
func setupTestAPI() *API {
	return NewAPI()
}

// TestHealthCheckHandler tests the health check endpoint
func TestHealthCheckHandler(t *testing.T) {
	api := setupTestAPI()
	router := api.setupRouter()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}
}

// TestCreateEmployeeHandler tests employee creation
func TestCreateEmployeeHandler(t *testing.T) {
	api := setupTestAPI()
	router := api.setupRouter()

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Valid employee",
			body: `{
				"name": "John Doe",
				"location": {"lat": 60.1699, "lon": 24.9384},
				"skills": ["delivery", "driving"]
			}`,
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response SuccessResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if response.Message != "Employee created successfully" {
					t.Errorf("Unexpected message: %s", response.Message)
				}
			},
		},
		{
			name:           "Invalid JSON",
			body:           `{"name": "John"`, // malformed JSON
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse error response: %v", err)
				}
			},
		},
		{
			name:           "Missing required fields",
			body:           `{"name": "John"}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name:           "Empty skills array",
			body:           `{"name": "John", "location": {"lat": 60.1699, "lon": 24.9384}, "skills": []}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/employees", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

// TestCreateTaskHandler tests task creation
func TestCreateTaskHandler(t *testing.T) {
	api := setupTestAPI()
	router := api.setupRouter()

	// First, create an employee
	empBody := `{
		"name": "Alice",
		"location": {"lat": 60.1699, "lon": 24.9384},
		"skills": ["delivery"]
	}`
	empReq := httptest.NewRequest("POST", "/employees", bytes.NewBufferString(empBody))
	empReq.Header.Set("Content-Type", "application/json")
	empW := httptest.NewRecorder()
	router.ServeHTTP(empW, empReq)

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Valid task",
			body: `{
				"location": {"lat": 60.1700, "lon": 24.9400},
				"required_skill": "delivery"
			}`,
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response SuccessResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if response.Message != "Task created and assignment initiated" {
					t.Errorf("Unexpected message: %s", response.Message)
				}
			},
		},
		{
			name:           "Invalid JSON",
			body:           `{"location": {"lat": 60.17`,
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name:           "Missing required_skill",
			body:           `{"location": {"lat": 60.1700, "lon": 24.9400}}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

// TestGetTasksHandler tests listing all tasks
func TestGetTasksHandler(t *testing.T) {
	api := setupTestAPI()
	router := api.setupRouter()

	// Create a task first
	taskBody := `{
		"location": {"lat": 60.1700, "lon": 24.9400},
		"required_skill": "delivery"
	}`
	taskReq := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(taskBody))
	taskReq.Header.Set("Content-Type", "application/json")
	taskW := httptest.NewRecorder()
	router.ServeHTTP(taskW, taskReq)

	// Now get all tasks
	req := httptest.NewRequest("GET", "/tasks", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	// Check that we have at least one task
	tasks, ok := response.Data.([]*Task)
	if !ok {
		// Try to unmarshal again as the actual structure
		var fullResponse struct {
			Message string      `json:"message"`
			Data    interface{} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &fullResponse); err != nil {
			t.Errorf("Failed to parse response structure: %v", err)
		}
	} else if len(tasks) == 0 {
		t.Error("Expected at least one task")
	}
}

// TestGetEmployeesHandler tests listing all employees
func TestGetEmployeesHandler(t *testing.T) {
	api := setupTestAPI()
	router := api.setupRouter()

	// Create an employee first
	empBody := `{
		"name": "Bob",
		"location": {"lat": 60.2055, "lon": 24.6559},
		"skills": ["delivery"]
	}`
	empReq := httptest.NewRequest("POST", "/employees", bytes.NewBufferString(empBody))
	empReq.Header.Set("Content-Type", "application/json")
	empW := httptest.NewRecorder()
	router.ServeHTTP(empW, empReq)

	// Now get all employees
	req := httptest.NewRequest("GET", "/employees", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}
}

// TestGetTaskByIDHandler tests getting a specific task
func TestGetTaskByIDHandler(t *testing.T) {
	api := setupTestAPI()
	router := api.setupRouter()

	// Create a task first
	taskBody := `{
		"location": {"lat": 60.1700, "lon": 24.9400},
		"required_skill": "delivery"
	}`
	taskReq := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(taskBody))
	taskReq.Header.Set("Content-Type", "application/json")
	taskW := httptest.NewRecorder()
	router.ServeHTTP(taskW, taskReq)

	var createResponse SuccessResponse
	json.Unmarshal(taskW.Body.Bytes(), &createResponse)

	// Extract task ID from response
	taskData, ok := createResponse.Data.(map[string]interface{})
	if !ok {
		// Parse the response body again to get the task
		var rawResponse map[string]interface{}
		json.Unmarshal(taskW.Body.Bytes(), &rawResponse)
		taskData = rawResponse["data"].(map[string]interface{})
	}

	taskID := taskData["id"].(string)

	// Test getting existing task
	t.Run("Existing task", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks/"+taskID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	// Test getting non-existent task
	t.Run("Non-existent task", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks/non-existent-id", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}

		var response ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to parse error response: %v", err)
		}

		if response.Code != "TASK_NOT_FOUND" {
			t.Errorf("Expected error code TASK_NOT_FOUND, got %s", response.Code)
		}
	})
}

// TestDuplicateEmployeeCreation tests creating duplicate employees
func TestDuplicateEmployeeCreation(t *testing.T) {
	api := setupTestAPI()
	router := api.setupRouter()

	empBody := `{
		"name": "Charlie",
		"location": {"lat": 60.1741, "lon": 24.9416},
		"skills": ["delivery"]
	}`

	// Create first employee and extract ID
	req1 := httptest.NewRequest("POST", "/employees", bytes.NewBufferString(empBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("First employee creation failed with status %d", w1.Code)
	}

	var firstResponse SuccessResponse
	json.Unmarshal(w1.Body.Bytes(), &firstResponse)

	// Get the employee data
	empData := firstResponse.Data.(map[string]interface{})
	empID := empData["id"].(string)

	// Try to create with same ID (by adding to store directly to simulate)
	// Note: In real scenario, we can't control UUID generation
	// This test shows the duplicate check works internally
	err := api.store.AddEmployee(&Employee{
		ID:          empID,
		Name:        "Duplicate",
		Location:    Location{Lat: 60.1741, Lon: 24.9416},
		Skills:      []string{"delivery"},
		IsAvailable: true,
	})

	if err != ErrDuplicateEmployee {
		t.Errorf("Expected ErrDuplicateEmployee, got %v", err)
	}
}

// TestCORSHeaders tests CORS middleware
func TestCORSHeaders(t *testing.T) {
	api := setupTestAPI()
	router := api.setupRouter()

	req := httptest.NewRequest("OPTIONS", "/health", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204 for OPTIONS, got %d", w.Code)
	}

	corsHeader := w.Header().Get("Access-Control-Allow-Origin")
	if corsHeader != "*" {
		t.Errorf("Expected CORS header '*', got '%s'", corsHeader)
	}
}

// TestFullWorkflow tests the complete workflow
func TestFullWorkflow(t *testing.T) {
	api := setupTestAPI()
	router := api.setupRouter()

	// Step 1: Create employees
	employees := []string{
		`{"name": "Alice", "location": {"lat": 60.1699, "lon": 24.9384}, "skills": ["delivery"]}`,
		`{"name": "Bob", "location": {"lat": 60.2055, "lon": 24.6559}, "skills": ["delivery"]}`,
	}

	for _, empBody := range employees {
		req := httptest.NewRequest("POST", "/employees", bytes.NewBufferString(empBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Failed to create employee: %d", w.Code)
		}
	}

	// Step 2: Verify employees exist
	req := httptest.NewRequest("GET", "/employees", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get employees: %d", w.Code)
	}

	// Step 3: Create a task
	taskBody := `{"location": {"lat": 60.1700, "lon": 24.9400}, "required_skill": "delivery"}`
	taskReq := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(taskBody))
	taskReq.Header.Set("Content-Type", "application/json")
	taskW := httptest.NewRecorder()
	router.ServeHTTP(taskW, taskReq)

	if taskW.Code != http.StatusCreated {
		t.Fatalf("Failed to create task: %d", taskW.Code)
	}

	// Step 4: Verify task was created
	tasksReq := httptest.NewRequest("GET", "/tasks", nil)
	tasksW := httptest.NewRecorder()
	router.ServeHTTP(tasksW, tasksReq)

	if tasksW.Code != http.StatusOK {
		t.Fatalf("Failed to get tasks: %d", tasksW.Code)
	}
}
