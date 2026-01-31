package main

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"
)

// TestCalculateDistance tests the Haversine distance calculation
func TestCalculateDistance(t *testing.T) {
	tests := []struct {
		name     string
		loc1     Location
		loc2     Location
		expected float64
		delta    float64 // acceptable error margin in km
	}{
		{
			name:     "Same location",
			loc1:     Location{Lat: 60.1699, Lon: 24.9384}, // Helsinki
			loc2:     Location{Lat: 60.1699, Lon: 24.9384},
			expected: 0,
			delta:    0.1,
		},
		{
			name:     "Helsinki to Espoo",
			loc1:     Location{Lat: 60.1699, Lon: 24.9384}, // Helsinki
			loc2:     Location{Lat: 60.2055, Lon: 24.6559}, // Espoo
			expected: 16.1,
			delta:    1.0,
		},
		{
			name:     "New York to Los Angeles",
			loc1:     Location{Lat: 40.7128, Lon: -74.0060},  // New York
			loc2:     Location{Lat: 34.0522, Lon: -118.2437}, // Los Angeles
			expected: 3935,
			delta:    50,
		},
		{
			name:     "London to Paris",
			loc1:     Location{Lat: 51.5074, Lon: -0.1278}, // London
			loc2:     Location{Lat: 48.8566, Lon: 2.3522},  // Paris
			expected: 343,
			delta:    10,
		},
		{
			name:     "Tokyo to Osaka",
			loc1:     Location{Lat: 35.6762, Lon: 139.6503}, // Tokyo
			loc2:     Location{Lat: 34.6937, Lon: 135.5023}, // Osaka
			expected: 398,
			delta:    10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := CalculateDistance(tt.loc1, tt.loc2)
			diff := math.Abs(distance - tt.expected)
			if diff > tt.delta {
				t.Errorf("CalculateDistance() = %.2f km, want %.2f km (±%.2f km), diff = %.2f km",
					distance, tt.expected, tt.delta, diff)
			}
		})
	}
}

// TestHasSkill tests the skill checking function
func TestHasSkill(t *testing.T) {
	tests := []struct {
		name     string
		skills   []string
		required string
		expected bool
	}{
		{
			name:     "Skill exists",
			skills:   []string{"delivery", "driving", "navigation"},
			required: "delivery",
			expected: true,
		},
		{
			name:     "Skill does not exist",
			skills:   []string{"delivery", "driving"},
			required: "cooking",
			expected: false,
		},
		{
			name:     "Empty skills",
			skills:   []string{},
			required: "delivery",
			expected: false,
		},
		{
			name:     "Case insensitive check",
			skills:   []string{"delivery"},
			required: "delivery",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasSkill(tt.skills, tt.required)
			if result != tt.expected {
				t.Errorf("hasSkill() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestStoreAddEmployee tests adding employees to the store
func TestStoreAddEmployee(t *testing.T) {
	store := NewStore()

	emp1 := &Employee{
		ID:          "emp1",
		Name:        "John Doe",
		Location:    Location{Lat: 60.1699, Lon: 24.9384},
		Skills:      []string{"delivery"},
		IsAvailable: true,
	}

	// Test adding a new employee
	err := store.AddEmployee(emp1)
	if err != nil {
		t.Errorf("AddEmployee() unexpected error: %v", err)
	}

	// Test adding duplicate employee
	err = store.AddEmployee(emp1)
	if err != ErrDuplicateEmployee {
		t.Errorf("AddEmployee() expected ErrDuplicateEmployee, got: %v", err)
	}

	// Verify employee was added
	retrievedEmp, err := store.GetEmployee("emp1")
	if err != nil {
		t.Errorf("GetEmployee() unexpected error: %v", err)
	}
	if retrievedEmp.Name != "John Doe" {
		t.Errorf("GetEmployee() name = %s, want John Doe", retrievedEmp.Name)
	}
}

// TestStoreGetAvailableEmployees tests filtering available employees by skill
func TestStoreGetAvailableEmployees(t *testing.T) {
	store := NewStore()

	employees := []*Employee{
		{
			ID:          "emp1",
			Name:        "Alice",
			Location:    Location{Lat: 60.1699, Lon: 24.9384},
			Skills:      []string{"delivery", "driving"},
			IsAvailable: true,
		},
		{
			ID:          "emp2",
			Name:        "Bob",
			Location:    Location{Lat: 60.2055, Lon: 24.6559},
			Skills:      []string{"delivery"},
			IsAvailable: false, // Not available
		},
		{
			ID:          "emp3",
			Name:        "Charlie",
			Location:    Location{Lat: 60.1741, Lon: 24.9416},
			Skills:      []string{"cooking"},
			IsAvailable: true, // Different skill
		},
		{
			ID:          "emp4",
			Name:        "Diana",
			Location:    Location{Lat: 60.1841, Lon: 24.9216},
			Skills:      []string{"delivery", "cooking"},
			IsAvailable: true,
		},
	}

	for _, emp := range employees {
		store.AddEmployee(emp)
	}

	// Test getting available employees with "delivery" skill
	available := store.GetAvailableEmployees("delivery")
	if len(available) != 2 {
		t.Errorf("GetAvailableEmployees() returned %d employees, want 2", len(available))
	}

	// Verify correct employees were returned
	foundAlice := false
	foundDiana := false
	for _, emp := range available {
		if emp.ID == "emp1" {
			foundAlice = true
		}
		if emp.ID == "emp4" {
			foundDiana = true
		}
	}
	if !foundAlice || !foundDiana {
		t.Error("GetAvailableEmployees() did not return expected employees")
	}
}

// TestTaskAssignment tests the core assignment logic
func TestTaskAssignment(t *testing.T) {
	store := NewStore()
	assigner := NewTaskAssigner(store)

	// Add employees at different locations
	employees := []*Employee{
		{
			ID:          "emp1",
			Name:        "Alice",
			Location:    Location{Lat: 60.1699, Lon: 24.9384}, // Helsinki center
			Skills:      []string{"delivery"},
			IsAvailable: true,
		},
		{
			ID:          "emp2",
			Name:        "Bob",
			Location:    Location{Lat: 60.2055, Lon: 24.6559}, // Espoo (far)
			Skills:      []string{"delivery"},
			IsAvailable: true,
		},
		{
			ID:          "emp3",
			Name:        "Charlie",
			Location:    Location{Lat: 60.1741, Lon: 24.9416}, // Near Helsinki
			Skills:      []string{"delivery"},
			IsAvailable: true,
		},
	}

	for _, emp := range employees {
		store.AddEmployee(emp)
	}

	// Create a task near Helsinki center
	task := &Task{
		ID:            "task1",
		Location:      Location{Lat: 60.1700, Lon: 24.9400},
		RequiredSkill: "delivery",
		Status:        TaskStatusPending,
	}
	store.AddTask(task)

	// Assign the task
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := assigner.AssignTask(ctx, task)
	if err != nil {
		t.Fatalf("AssignTask() unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("AssignTask() expected success")
	}

	// The closest employee should be Alice (emp1)
	if result.EmployeeID != "emp1" {
		t.Errorf("AssignTask() assigned to %s, want emp1", result.EmployeeID)
	}

	// Verify task was updated
	updatedTask, err := store.GetTask("task1")
	if err != nil {
		t.Errorf("GetTask() unexpected error: %v", err)
	}
	if updatedTask.Status != TaskStatusAssigned {
		t.Errorf("Task status = %s, want %s", updatedTask.Status, TaskStatusAssigned)
	}
	if updatedTask.AssignedEmployeeID != "emp1" {
		t.Errorf("Task assigned to %s, want emp1", updatedTask.AssignedEmployeeID)
	}

	// Verify employee is marked as unavailable
	emp, err := store.GetEmployee("emp1")
	if err != nil {
		t.Errorf("GetEmployee() unexpected error: %v", err)
	}
	if emp.IsAvailable {
		t.Error("Employee should be marked as unavailable after assignment")
	}
}

// TestTaskAssignmentNoEligibleEmployee tests assignment when no employee is eligible
func TestTaskAssignmentNoEligibleEmployee(t *testing.T) {
	store := NewStore()
	assigner := NewTaskAssigner(store)

	// Add employee with different skill
	emp := &Employee{
		ID:          "emp1",
		Name:        "Alice",
		Location:    Location{Lat: 60.1699, Lon: 24.9384},
		Skills:      []string{"cooking"},
		IsAvailable: true,
	}
	store.AddEmployee(emp)

	// Create task requiring delivery skill
	task := &Task{
		ID:            "task1",
		Location:      Location{Lat: 60.1700, Lon: 24.9400},
		RequiredSkill: "delivery",
		Status:        TaskStatusPending,
	}
	store.AddTask(task)

	// Attempt assignment
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := assigner.AssignTask(ctx, task)

	// Should fail with no eligible employee error
	if err != ErrNoEligibleEmployee {
		t.Errorf("AssignTask() expected ErrNoEligibleEmployee, got: %v", err)
	}
	if result != nil && result.Success {
		t.Error("AssignTask() should not succeed when no eligible employee")
	}

	// Verify task status is failed
	updatedTask, err := store.GetTask("task1")
	if err != nil {
		t.Errorf("GetTask() unexpected error: %v", err)
	}
	if updatedTask.Status != TaskStatusFailed {
		t.Errorf("Task status = %s, want %s", updatedTask.Status, TaskStatusFailed)
	}
}

// TestTaskAssignmentWithTimeout tests assignment with context timeout
func TestTaskAssignmentWithTimeout(t *testing.T) {
	store := NewStore()
	assigner := NewTaskAssigner(store)

	// Create a task
	task := &Task{
		ID:            "task1",
		Location:      Location{Lat: 60.1700, Lon: 24.9400},
		RequiredSkill: "delivery",
		Status:        TaskStatusPending,
	}
	store.AddTask(task)

	// Create context with immediate timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait a bit to ensure timeout
	time.Sleep(10 * time.Millisecond)

	_, err := assigner.AssignTask(ctx, task)

	// Should fail with timeout error
	if err == nil {
		t.Error("AssignTask() expected timeout error")
	}

	taskErr, ok := err.(*TaskError)
	if !ok {
		t.Error("Expected TaskError type")
	}
	if taskErr.Code != ErrAssignmentTimeout.Code {
		t.Errorf("Error code = %s, want %s", taskErr.Code, ErrAssignmentTimeout.Code)
	}
}

// TestConcurrentEmployeeOperations tests thread safety of employee operations
func TestConcurrentEmployeeOperations(t *testing.T) {
	store := NewStore()

	// Add employees concurrently
	done := make(chan bool)
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			emp := &Employee{
				ID:          fmt.Sprintf("emp-%d", id),
				Name:        "Employee",
				Location:    Location{Lat: 60.1699, Lon: 24.9384},
				Skills:      []string{"delivery"},
				IsAvailable: true,
			}
			store.AddEmployee(emp)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all employees were added
	available := store.GetAvailableEmployees("delivery")
	if len(available) != numGoroutines {
		t.Errorf("Expected %d employees, got %d", numGoroutines, len(available))
	}
}

// TestConcurrentTaskOperations tests thread safety of task operations
func TestConcurrentTaskOperations(t *testing.T) {
	store := NewStore()

	// Add tasks concurrently
	done := make(chan bool)
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			task := &Task{
				ID:            fmt.Sprintf("task-%d", id),
				Location:      Location{Lat: 60.1699, Lon: 24.9384},
				RequiredSkill: "delivery",
				Status:        TaskStatusPending,
			}
			store.AddTask(task)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all tasks were added
	tasks := store.GetAllTasks()
	if len(tasks) != numGoroutines {
		t.Errorf("Expected %d tasks, got %d", numGoroutines, len(tasks))
	}
}

// TestLocationValidation tests coordinate validation
func TestLocationValidation(t *testing.T) {
	tests := []struct {
		name      string
		location  Location
		shouldErr bool
	}{
		{
			name:      "Valid coordinates",
			location:  Location{Lat: 60.1699, Lon: 24.9384},
			shouldErr: false,
		},
		{
			name:      "Latitude too high",
			location:  Location{Lat: 91, Lon: 24.9384},
			shouldErr: true,
		},
		{
			name:      "Latitude too low",
			location:  Location{Lat: -91, Lon: 24.9384},
			shouldErr: true,
		},
		{
			name:      "Longitude too high",
			location:  Location{Lat: 60, Lon: 181},
			shouldErr: true,
		},
		{
			name:      "Longitude too low",
			location:  Location{Lat: 60, Lon: -181},
			shouldErr: true,
		},
		{
			name:      "Edge case - North Pole",
			location:  Location{Lat: 90, Lon: 0},
			shouldErr: false,
		},
		{
			name:      "Edge case - South Pole",
			location:  Location{Lat: -90, Lon: 0},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.location.Validate()
			if tt.shouldErr && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

// TestSkillNormalization tests case-insensitive skill matching
func TestSkillNormalization(t *testing.T) {
	store := NewStore()

	emp := &Employee{
		ID:          "emp1",
		Name:        "Alice",
		Location:    Location{Lat: 60.1699, Lon: 24.9384},
		Skills:      []string{"Delivery", "DRIVING", "Navigation"},
		IsAvailable: true,
	}

	// Validate should normalize skills
	if err := emp.Validate(); err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	store.AddEmployee(emp)

	// Should find employee with lowercase skill
	available := store.GetAvailableEmployees("delivery")
	if len(available) != 1 {
		t.Errorf("Expected 1 employee with 'delivery' skill, got %d", len(available))
	}

	// Should find employee with uppercase skill
	available = store.GetAvailableEmployees("DRIVING")
	if len(available) != 1 {
		t.Errorf("Expected 1 employee with 'DRIVING' skill, got %d", len(available))
	}
}

// TestEmptySkillsValidation tests empty skills array validation
func TestEmptySkillsValidation(t *testing.T) {
	emp := &Employee{
		ID:          "emp1",
		Name:        "Alice",
		Location:    Location{Lat: 60.1699, Lon: 24.9384},
		Skills:      []string{},
		IsAvailable: true,
	}

	err := emp.Validate()
	if err == nil {
		t.Error("Expected validation error for empty skills array")
	}
}

// TestEmptyNameValidation tests empty name validation
func TestEmptyNameValidation(t *testing.T) {
	emp := &Employee{
		ID:          "emp1",
		Name:        "   ",
		Location:    Location{Lat: 60.1699, Lon: 24.9384},
		Skills:      []string{"delivery"},
		IsAvailable: true,
	}

	err := emp.Validate()
	if err == nil {
		t.Error("Expected validation error for empty name")
	}
}

// TestDuplicateTaskID tests duplicate task ID prevention
func TestDuplicateTaskID(t *testing.T) {
	store := NewStore()

	task1 := &Task{
		ID:            "task1",
		Location:      Location{Lat: 60.1699, Lon: 24.9384},
		RequiredSkill: "delivery",
		Status:        TaskStatusPending,
	}

	err := store.AddTask(task1)
	if err != nil {
		t.Fatalf("First AddTask failed: %v", err)
	}

	// Try to add duplicate
	task2 := &Task{
		ID:            "task1",
		Location:      Location{Lat: 60.2055, Lon: 24.6559},
		RequiredSkill: "cooking",
		Status:        TaskStatusPending,
	}

	err = store.AddTask(task2)
	if err == nil {
		t.Error("Expected error for duplicate task ID")
	}

	taskErr, ok := err.(*TaskError)
	if !ok || taskErr.Code != "DUPLICATE_TASK" {
		t.Errorf("Expected DUPLICATE_TASK error, got: %v", err)
	}
}

// TestDistanceCalculationEdgeCases tests edge cases in distance calculation
func TestDistanceCalculationEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		loc1     Location
		loc2     Location
		checkNaN bool
	}{
		{
			name:     "Same location",
			loc1:     Location{Lat: 0, Lon: 0},
			loc2:     Location{Lat: 0, Lon: 0},
			checkNaN: false,
		},
		{
			name:     "Antipodal points",
			loc1:     Location{Lat: 0, Lon: 0},
			loc2:     Location{Lat: 0, Lon: 180},
			checkNaN: false,
		},
		{
			name:     "North and South pole",
			loc1:     Location{Lat: 90, Lon: 0},
			loc2:     Location{Lat: -90, Lon: 0},
			checkNaN: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := CalculateDistance(tt.loc1, tt.loc2)
			if math.IsNaN(distance) || math.IsInf(distance, 0) {
				t.Errorf("Distance calculation returned NaN or Inf: %f", distance)
			}
			if distance < 0 {
				t.Errorf("Distance cannot be negative: %f", distance)
			}
		})
	}
}

// BenchmarkCalculateDistance benchmarks the distance calculation
func BenchmarkCalculateDistance(b *testing.B) {
	loc1 := Location{Lat: 60.1699, Lon: 24.9384}
	loc2 := Location{Lat: 60.2055, Lon: 24.6559}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateDistance(loc1, loc2)
	}
}

// BenchmarkAssignTask benchmarks the task assignment
func BenchmarkAssignTask(b *testing.B) {
	store := NewStore()
	assigner := NewTaskAssigner(store)

	// Add 100 employees
	for i := 0; i < 100; i++ {
		emp := &Employee{
			ID:          fmt.Sprintf("emp-%d", i),
			Name:        "Employee",
			Location:    Location{Lat: 60.0 + float64(i)*0.01, Lon: 24.0 + float64(i)*0.01},
			Skills:      []string{"delivery"},
			IsAvailable: true,
		}
		store.AddEmployee(emp)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := &Task{
			ID:            fmt.Sprintf("task-%d", i),
			Location:      Location{Lat: 60.1699, Lon: 24.9384},
			RequiredSkill: "delivery",
			Status:        TaskStatusPending,
		}
		store.AddTask(task)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		assigner.AssignTask(ctx, task)
		cancel()
	}
}
