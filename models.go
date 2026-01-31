package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// Location represents geographical coordinates
type Location struct {
	Lat float64 `json:"lat" binding:"required"`
	Lon float64 `json:"lon" binding:"required"`
}

// Validate checks if coordinates are within valid ranges
func (l Location) Validate() error {
	if l.Lat < -90 || l.Lat > 90 {
		return fmt.Errorf("latitude must be between -90 and 90, got %.6f", l.Lat)
	}
	if l.Lon < -180 || l.Lon > 180 {
		return fmt.Errorf("longitude must be between -180 and 180, got %.6f", l.Lon)
	}
	return nil
}

// normalizeSkill converts skill to lowercase for case-insensitive matching
func normalizeSkill(skill string) string {
	return strings.ToLower(strings.TrimSpace(skill))
}

// normalizeSkills normalizes all skills in a slice
func normalizeSkills(skills []string) []string {
	normalized := make([]string, len(skills))
	for i, skill := range skills {
		normalized[i] = normalizeSkill(skill)
	}
	return normalized
}

// validateSkills checks if skills array is valid
func validateSkills(skills []string) error {
	if len(skills) == 0 {
		return errors.New("skills array cannot be empty")
	}
	for i, skill := range skills {
		if strings.TrimSpace(skill) == "" {
			return fmt.Errorf("skill at index %d cannot be empty or whitespace", i)
		}
	}
	return nil
}

// Employee represents a worker who can be assigned tasks
type Employee struct {
	ID          string   `json:"id" binding:"required"`
	Name        string   `json:"name" binding:"required"`
	Location    Location `json:"location" binding:"required"`
	Skills      []string `json:"skills" binding:"required"`
	IsAvailable bool     `json:"is_available"`
}

// Validate validates employee data
func (e *Employee) Validate() error {
	if strings.TrimSpace(e.Name) == "" {
		return errors.New("employee name cannot be empty")
	}
	if err := e.Location.Validate(); err != nil {
		return fmt.Errorf("invalid location: %w", err)
	}
	if err := validateSkills(e.Skills); err != nil {
		return fmt.Errorf("invalid skills: %w", err)
	}
	// Normalize skills for case-insensitive comparison
	e.Skills = normalizeSkills(e.Skills)
	return nil
}

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	TaskStatusPending  TaskStatus = "pending"
	TaskStatusAssigned TaskStatus = "assigned"
	TaskStatusFailed   TaskStatus = "failed"
)

// Task represents a job that needs to be assigned to an employee
type Task struct {
	ID                 string     `json:"id" binding:"required"`
	Location           Location   `json:"location" binding:"required"`
	RequiredSkill      string     `json:"required_skill" binding:"required"`
	Status             TaskStatus `json:"status"`
	AssignedEmployeeID string     `json:"assigned_employee_id,omitempty"`
}

// Validate validates task data
func (t *Task) Validate() error {
	if err := t.Location.Validate(); err != nil {
		return fmt.Errorf("invalid location: %w", err)
	}
	if strings.TrimSpace(t.RequiredSkill) == "" {
		return errors.New("required_skill cannot be empty")
	}
	// Normalize skill for case-insensitive comparison
	t.RequiredSkill = normalizeSkill(t.RequiredSkill)
	return nil
}

// Custom error types for better error handling
type TaskError struct {
	Code    string
	Message string
	Err     error
}

func (e *TaskError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *TaskError) Unwrap() error {
	return e.Err
}

// Error codes
var (
	ErrNoEligibleEmployee = &TaskError{
		Code:    "NO_ELIGIBLE_EMPLOYEE",
		Message: "No eligible employee found for task assignment",
	}
	ErrEmployeeNotFound = &TaskError{
		Code:    "EMPLOYEE_NOT_FOUND",
		Message: "Employee not found",
	}
	ErrTaskNotFound = &TaskError{
		Code:    "TASK_NOT_FOUND",
		Message: "Task not found",
	}
	ErrDuplicateEmployee = &TaskError{
		Code:    "DUPLICATE_EMPLOYEE",
		Message: "Employee with this ID already exists",
	}
	ErrAssignmentTimeout = &TaskError{
		Code:    "ASSIGNMENT_TIMEOUT",
		Message: "Task assignment timed out",
	}
	ErrEmployeeNoLongerAvailable = &TaskError{
		Code:    "EMPLOYEE_UNAVAILABLE",
		Message: "Selected employee no longer available (assigned concurrently)",
	}
)

// Store provides thread-safe in-memory storage for employees and tasks
type Store struct {
	employees map[string]*Employee
	tasks     map[string]*Task
	mu        sync.RWMutex
}

// NewStore creates a new Store instance
func NewStore() *Store {
	return &Store{
		employees: make(map[string]*Employee),
		tasks:     make(map[string]*Task),
	}
}

// AddEmployee adds a new employee to the store
func (s *Store) AddEmployee(emp *Employee) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.employees[emp.ID]; exists {
		return ErrDuplicateEmployee
	}

	s.employees[emp.ID] = emp
	return nil
}

// GetEmployee retrieves an employee by ID
func (s *Store) GetEmployee(id string) (*Employee, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	emp, exists := s.employees[id]
	if !exists {
		return nil, ErrEmployeeNotFound
	}
	return emp, nil
}

// GetAvailableEmployees returns all available employees with a specific skill
func (s *Store) GetAvailableEmployees(skill string) []*Employee {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var eligible []*Employee
	for _, emp := range s.employees {
		if emp.IsAvailable && hasSkill(emp.Skills, skill) {
			eligible = append(eligible, emp)
		}
	}
	return eligible
}

// UpdateEmployeeAvailability updates an employee's availability status
func (s *Store) UpdateEmployeeAvailability(id string, available bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	emp, exists := s.employees[id]
	if !exists {
		return ErrEmployeeNotFound
	}
	emp.IsAvailable = available
	return nil
}

// AddTask adds a new task to the store
func (s *Store) AddTask(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[task.ID]; exists {
		return &TaskError{
			Code:    "DUPLICATE_TASK",
			Message: "Task with this ID already exists",
		}
	}

	task.Status = TaskStatusPending
	s.tasks[task.ID] = task
	return nil
}

// GetTask retrieves a task by ID
func (s *Store) GetTask(id string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

// GetAllTasks returns all tasks
func (s *Store) GetAllTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// UpdateTask updates a task's status and assignment
func (s *Store) UpdateTask(id string, status TaskStatus, employeeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return ErrTaskNotFound
	}
	task.Status = status
	task.AssignedEmployeeID = employeeID
	return nil
}

// hasSkill checks if an employee has a specific skill (case-insensitive)
// NOTE: skills are already normalized when employee is created
func hasSkill(skills []string, required string) bool {
	requiredNorm := normalizeSkill(required)
	for _, skill := range skills {
		// Skills are pre-normalized, direct comparison
		if skill == requiredNorm {
			return true
		}
	}
	return false
}

// CalculateDistance calculates the distance between two locations using the Haversine formula
// Returns distance in kilometers with numerical stability checks
func CalculateDistance(loc1, loc2 Location) float64 {
	const earthRadiusKm = 6371.0

	// Convert degrees to radians
	lat1Rad := loc1.Lat * math.Pi / 180
	lat2Rad := loc2.Lat * math.Pi / 180
	deltaLat := (loc2.Lat - loc1.Lat) * math.Pi / 180
	deltaLon := (loc2.Lon - loc1.Lon) * math.Pi / 180

	// Haversine formula with stability improvements
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	// Clamp 'a' to prevent floating-point errors (should be in [0,1])
	a = math.Max(0, math.Min(1, a))

	// Protect against numerical instability for sqrt
	sqrtA := math.Sqrt(a)
	sqrt1MinusA := math.Sqrt(1 - a)

	// Check for NaN or Inf
	if math.IsNaN(sqrtA) || math.IsInf(sqrtA, 0) || math.IsNaN(sqrt1MinusA) || math.IsInf(sqrt1MinusA, 0) {
		return 0 // Return 0 for invalid calculations
	}

	c := 2 * math.Atan2(sqrtA, sqrt1MinusA)
	distance := earthRadiusKm * c

	// Final safety check
	if math.IsNaN(distance) || math.IsInf(distance, 0) {
		return 0
	}

	return distance
}

// AssignmentResult represents the result of a task assignment attempt
type AssignmentResult struct {
	TaskID     string
	EmployeeID string
	Distance   float64
	Success    bool
	Error      error
}

// TaskAssigner handles the assignment of tasks to employees
type TaskAssigner struct {
	store *Store
}

// NewTaskAssigner creates a new TaskAssigner
func NewTaskAssigner(store *Store) *TaskAssigner {
	return &TaskAssigner{store: store}
}

// AssignTask assigns a task to the closest eligible employee
// Uses context for timeout management
// NOTE: Caller is responsible for running in goroutine if async behavior is needed
func (ta *TaskAssigner) AssignTask(ctx context.Context, task *Task) (*AssignmentResult, error) {
	// Check context deadline before attempting assignment
	select {
	case <-ctx.Done():
		// Context already cancelled/timed out
		ta.store.mu.Lock()
		if t, exists := ta.store.tasks[task.ID]; exists {
			t.Status = TaskStatusFailed
		}
		ta.store.mu.Unlock()
		return nil, &TaskError{
			Code:    ErrAssignmentTimeout.Code,
			Message: ErrAssignmentTimeout.Message,
			Err:     ctx.Err(),
		}
	default:
	}

	// Perform assignment directly (no goroutine)
	return ta.performAssignment(ctx, task)
}

// performAssignment performs the actual assignment logic with two-phase locking
// Phase 1: Read employees under RLock
// Phase 2: Calculate distances without lock (CPU-bound work)
// Phase 3: Atomic compare-and-swap under Lock
func (ta *TaskAssigner) performAssignment(ctx context.Context, task *Task) (*AssignmentResult, error) {
	// Phase 1: Snapshot eligible employees under read lock
	ta.store.mu.RLock()
	type employeeSnapshot struct {
		id       string
		location Location
	}
	var eligible []employeeSnapshot
	for _, emp := range ta.store.employees {
		if emp.IsAvailable && hasSkill(emp.Skills, task.RequiredSkill) {
			eligible = append(eligible, employeeSnapshot{
				id:       emp.ID,
				location: emp.Location,
			})
		}
	}
	ta.store.mu.RUnlock()

	if len(eligible) == 0 {
		// No eligible employees, mark task as failed
		ta.store.mu.Lock()
		if t, exists := ta.store.tasks[task.ID]; exists {
			t.Status = TaskStatusFailed
			t.AssignedEmployeeID = ""
		}
		ta.store.mu.Unlock()
		return &AssignmentResult{
			TaskID:  task.ID,
			Success: false,
			Error:   ErrNoEligibleEmployee,
		}, ErrNoEligibleEmployee
	}

	// Phase 2: Calculate distances WITHOUT holding lock (expensive CPU work)
	// BUT check context periodically to avoid wasted work
	var closestID string
	minDistance := math.MaxFloat64
	for i, emp := range eligible {
		// Check context every 10 employees to catch cancellation
		if i%10 == 0 {
			select {
			case <-ctx.Done():
				// Context cancelled during calculation, fail immediately
				ta.store.mu.Lock()
				if t, exists := ta.store.tasks[task.ID]; exists {
					t.Status = TaskStatusFailed
				}
				ta.store.mu.Unlock()
				return nil, &TaskError{
					Code:    ErrAssignmentTimeout.Code,
					Message: ErrAssignmentTimeout.Message,
					Err:     ctx.Err(),
				}
			default:
			}
		}
		distance := CalculateDistance(task.Location, emp.location)
		if distance < minDistance {
			minDistance = distance
			closestID = emp.id
		}
	}

	// Phase 3: Atomic CAS - re-check availability and assign
	ta.store.mu.Lock()
	defer ta.store.mu.Unlock()

	// Final context check before committing assignment
	select {
	case <-ctx.Done():
		if t, exists := ta.store.tasks[task.ID]; exists {
			t.Status = TaskStatusFailed
		}
		return nil, &TaskError{
			Code:    ErrAssignmentTimeout.Code,
			Message: ErrAssignmentTimeout.Message,
			Err:     ctx.Err(),
		}
	default:
	}

	// Re-check that the closest employee is still available (CAS)
	emp, exists := ta.store.employees[closestID]
	if !exists || !emp.IsAvailable {
		// Employee was assigned to another task concurrently
		// This is NOT "no eligible employee" - it's a CAS race condition
		if t, exists := ta.store.tasks[task.ID]; exists {
			t.Status = TaskStatusFailed
			t.AssignedEmployeeID = ""
		}
		return &AssignmentResult{
			TaskID:  task.ID,
			Success: false,
			Error:   ErrEmployeeNoLongerAvailable,
		}, ErrEmployeeNoLongerAvailable
	}

	// Atomically assign task and mark employee unavailable
	emp.IsAvailable = false
	if t, exists := ta.store.tasks[task.ID]; exists {
		t.Status = TaskStatusAssigned
		t.AssignedEmployeeID = closestID
	}

	return &AssignmentResult{
		TaskID:     task.ID,
		EmployeeID: closestID,
		Distance:   minDistance,
		Success:    true,
	}, nil
}

// AssignmentWorkerPool manages concurrent task assignments
type AssignmentWorkerPool struct {
	assigner   *TaskAssigner
	taskQueue  chan *Task
	numWorkers int
	timeout    time.Duration
	wg         sync.WaitGroup
}

// NewAssignmentWorkerPool creates a new worker pool
func NewAssignmentWorkerPool(assigner *TaskAssigner, numWorkers int, timeout time.Duration) *AssignmentWorkerPool {
	return &AssignmentWorkerPool{
		assigner:   assigner,
		taskQueue:  make(chan *Task, 100),
		numWorkers: numWorkers,
		timeout:    timeout,
	}
}

// Start starts the worker pool
func (pool *AssignmentWorkerPool) Start(ctx context.Context) {
	for i := 0; i < pool.numWorkers; i++ {
		pool.wg.Add(1)
		go pool.worker(ctx, i)
	}
}

// worker processes tasks from the queue
// Shutdown is triggered by closing taskQueue channel (not context)
// Context is only used for per-task timeouts
func (pool *AssignmentWorkerPool) worker(ctx context.Context, workerID int) {
	defer pool.wg.Done()

	// Single shutdown mechanism: closed channel
	for task := range pool.taskQueue {
		// Nil-safety: should never happen, but defensive check
		if task == nil {
			fmt.Printf("Worker %d: Received nil task, skipping\n", workerID)
			continue
		}

		// Check if shutdown context is cancelled (for graceful drain)
		select {
		case <-ctx.Done():
			// Context cancelled but channel not closed yet
			// Fail remaining tasks quickly
			fmt.Printf("Worker %d: Context cancelled, failing task %s\n", workerID, task.ID)
			pool.assigner.store.mu.Lock()
			if t, exists := pool.assigner.store.tasks[task.ID]; exists {
				t.Status = TaskStatusFailed
			}
			pool.assigner.store.mu.Unlock()
			continue
		default:
		}

		// Normal processing with per-task timeout
		assignCtx, cancel := context.WithTimeout(ctx, pool.timeout)
		_, err := pool.assigner.AssignTask(assignCtx, task)
		if err != nil {
			fmt.Printf("Worker %d: Failed to assign task %s: %v\n", workerID, task.ID, err)
		} else {
			fmt.Printf("Worker %d: Successfully assigned task %s\n", workerID, task.ID)
		}
		cancel()
	}

	fmt.Printf("Worker %d: Queue closed, exiting\n", workerID)
}

// SubmitTask submits a task to the worker pool (non-blocking)
// Returns error if queue is full
func (pool *AssignmentWorkerPool) SubmitTask(task *Task) error {
	select {
	case pool.taskQueue <- task:
		return nil
	default:
		return &TaskError{
			Code:    "QUEUE_FULL",
			Message: "Worker pool queue is full, please try again later",
		}
	}
}

// Shutdown gracefully shuts down the worker pool
// Closes the queue and waits for all workers to finish
func (pool *AssignmentWorkerPool) Shutdown() {
	close(pool.taskQueue)
	pool.wg.Wait()
}
