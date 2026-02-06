# Task Assignment Engine - Architecture Diagram

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        CLIENT (curl, Postman, App)                  │
└────────────────────────────────┬────────────────────────────────────┘
                                 │ HTTP/JSON
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          GIN WEB FRAMEWORK                           │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  Routes & Middleware                                          │  │
│  │  • POST /employees     • GET /employees                       │  │
│  │  • POST /tasks         • GET /tasks                           │  │
│  │  • GET /tasks/:id      • GET /health                          │  │
│  │  • CORS, Error Handling, Logging                              │  │
│  └──────────────────────────────────────────────────────────────┘  │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          API LAYER (main.go)                         │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  Request Handlers                                             │  │
│  │  • handleCreateEmployee()    • handleGetTasks()               │  │
│  │  • handleCreateTask()        • handleGetEmployees()           │  │
│  │  • handleHealthCheck()                                        │  │
│  └────────────┬─────────────────────────────────────────────────┘  │
└───────────────┼────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    BUSINESS LOGIC LAYER (models.go)                 │
│                                                                     │
│  ┌────────────────────────────────────────────────────────────┐     │
│  │  TaskAssigner (Core Algorithm)                             │    │
│  │  ┌──────────────────────────────────────────────────────┐  │    │
│  │  │  1. Filter employees by skill & availability         │  │    │
│  │  │  2. Calculate distances (Haversine)                  │  │    │
│  │  │  3. Select closest employee                          │  │    │
│  │  │  4. Update task & employee status                    │  │    │
│  │  └──────────────────────────────────────────────────────┘  │    │
│  └────────────────────────────────────────────────────────────┘    │
│                                                                    │
│  ┌────────────────────────────────────────────────────────────┐    │
│  │  AssignmentWorkerPool (Concurrency)                         │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │    │
│  │  │ Worker 1 │  │ Worker 2 │  │ Worker 3 │  │ Worker 4 │  │    │
│  │  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  │    │
│  │       │             │             │             │          │    │
│  │       └─────────────┴─────────────┴─────────────┘          │    │
│  │                      │                                      │    │
│  │            ┌─────────▼─────────┐                           │    │
│  │            │   Task Queue      │                           │    │
│  │            │  (Channel: 100)   │                           │    │
│  │            └───────────────────┘                           │    │
│  └────────────────────────────────────────────────────────────┘    │
│                                                                       │
│  ┌────────────────────────────────────────────────────────────┐    │
│  │  Helper Functions                                           │    │
│  │  • CalculateDistance() - Haversine formula                  │    │
│  │  • hasSkill() - Skill matching                              │    │
│  └────────────────────────────────────────────────────────────┘    │
└─────────────────────────────┬───────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     DATA LAYER (In-Memory Store)                     │
│  ┌────────────────────────────────────────────────────────────┐    │
│  │  Thread-Safe Store (sync.RWMutex)                           │    │
│  │  ┌──────────────────────────┐  ┌──────────────────────┐   │    │
│  │  │  Employee Map            │  │  Task Map            │   │    │
│  │  │  map[string]*Employee    │  │  map[string]*Task    │   │    │
│  │  │                          │  │                      │   │    │
│  │  │  Key: Employee ID        │  │  Key: Task ID        │   │    │
│  │  │  Value: Employee object  │  │  Value: Task object  │   │    │
│  │  └──────────────────────────┘  └──────────────────────┘   │    │
│  │                                                              │    │
│  │  Operations: CRUD with mutex locking                        │    │
│  └────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────┘
```

## Data Flow: Task Assignment

```
┌──────────────┐
│  Client      │
│  POST /tasks │
└──────┬───────┘
       │
       │ 1. Create task request
       ▼
┌──────────────────┐
│  API Handler     │
│  handleCreateTask│
└──────┬───────────┘
       │
       │ 2. Generate UUID, create Task object
       ▼
┌──────────────────┐
│  Store           │
│  AddTask()       │◄────────────────┐
└──────┬───────────┘                 │
       │                             │
       │ 3. Task stored              │
       ▼                             │
┌──────────────────┐                 │
│  Worker Pool     │                 │
│  SubmitTask()    │                 │
└──────┬───────────┘                 │
       │                             │
       │ 4. Task submitted to channel│
       ▼                             │
┌──────────────────┐                 │
│  Worker Goroutine│                 │
│  (1 of 5)        │                 │
└──────┬───────────┘                 │
       │                             │
       │ 5. Process task             │
       ▼                             │
┌──────────────────────────┐         │
│  TaskAssigner            │         │
│  AssignTask()            │         │
│  (with context timeout)  │         │
└──────┬───────────────────┘         │
       │                             │
       │ 6. Get available employees  │
       ▼                             │
┌──────────────────────────┐         │
│  Store                   │         │
│  GetAvailableEmployees() │         │
└──────┬───────────────────┘         │
       │                             │
       │ 7. Filter by skill          │
       ▼                             │
┌──────────────────────────┐         │
│  hasSkill()              │         │
│  (for each employee)     │         │
└──────┬───────────────────┘         │
       │                             │
       │ 8. Calculate distances      │
       ▼                             │
┌──────────────────────────┐         │
│  CalculateDistance()     │         │
│  (Haversine formula)     │         │
└──────┬───────────────────┘         │
       │                             │
       │ 9. Find closest employee    │
       ▼                             │
┌──────────────────────────┐         │
│  Select minimum distance │         │
└──────┬───────────────────┘         │
       │                             │
       │ 10. Update task status      │
       │     Update employee         │
       └─────────────────────────────┘
```

## Concurrency Model

```
┌────────────────────────────────────────────────────────────────┐
│                     MAIN GOROUTINE                              │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  HTTP Server                                              │  │
│  │  • Listens on :8080                                       │  │
│  │  • Handles incoming requests                              │  │
│  │  • Returns responses immediately                          │  │
│  └──────────────────────────────────────────────────────────┘  │
└───────────────────────────┬────────────────────────────────────┘
                            │
                            │ Task submission (non-blocking)
                            ▼
┌────────────────────────────────────────────────────────────────┐
│                     TASK QUEUE (Channel)                        │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  chan *Task (buffered, capacity: 100)                     │  │
│  └──────────────────────────────────────────────────────────┘  │
└──┬────────┬────────┬────────┬────────┬────────────────────────┘
   │        │        │        │        │
   │        │        │        │        │
   ▼        ▼        ▼        ▼        ▼
┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐
│Worker│ │Worker│ │Worker│ │Worker│ │Worker│
│  1   │ │  2   │ │  3   │ │  4   │ │  5   │
└──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘
   │        │        │        │        │
   │ Each worker:                      │
   │ 1. Reads from channel             │
   │ 2. Creates context with timeout   │
   │ 3. Calls AssignTask()             │
   │ 4. Logs result                    │
   │ 5. Repeats                        │
   └───────┴────────┴────────┴─────────┘
```

## Thread Safety Mechanisms

```
┌─────────────────────────────────────────────────────────────┐
│                      STORE (Thread-Safe)                     │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  sync.RWMutex                                           │ │
│  │  ┌────────────────────┐  ┌────────────────────┐        │ │
│  │  │  Read Operations   │  │  Write Operations  │        │ │
│  │  │  (RLock/RUnlock)   │  │  (Lock/Unlock)     │        │ │
│  │  │                    │  │                    │        │ │
│  │  │  • GetEmployee     │  │  • AddEmployee     │        │ │
│  │  │  • GetTask         │  │  • AddTask         │        │ │
│  │  │  • GetAllTasks     │  │  • UpdateTask      │        │ │
│  │  │  • GetAvailable    │  │  • UpdateAvail     │        │ │
│  │  │    Employees       │  │                    │        │ │
│  │  └────────────────────┘  └────────────────────┘        │ │
│  │                                                          │ │
│  │  Multiple readers OR single writer                      │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Error Handling Flow

```
┌──────────────────────────────────────────────────────────────┐
│                    CUSTOM ERROR TYPES                         │
│                                                               │
│  type TaskError struct {                                     │
│      Code    string                                          │
│      Message string                                          │
│      Err     error                                           │
│  }                                                            │
│                                                               │
│  Predefined Errors:                                          │
│  • ErrNoEligibleEmployee   → HTTP 404                        │
│  • ErrEmployeeNotFound     → HTTP 404                        │
│  • ErrTaskNotFound         → HTTP 404                        │
│  • ErrDuplicateEmployee    → HTTP 409                        │
│  • ErrAssignmentTimeout    → HTTP 408                        │
└──────────────────────────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────┐
│                      ERROR PROPAGATION                        │
│                                                               │
│  Business Logic → TaskError                                  │
│        ↓                                                      │
│  API Handler → HTTP Status Code + JSON Response              │
│        ↓                                                      │
│  Client ← {error, code, message}                             │
└──────────────────────────────────────────────────────────────┘
```

## Context and Timeout Management

```
┌──────────────────────────────────────────────────────────────┐
│                    CONTEXT FLOW                               │
│                                                               │
│  1. Worker creates context with 30s timeout                  │
│     ctx, cancel := context.WithTimeout(bg, 30*time.Second)  │
│                                                               │
│  2. Context passed to AssignTask()                           │
│     result, err := assigner.AssignTask(ctx, task)           │
│                                                               │
│  3. AssignTask runs in goroutine                            │
│     • Performs assignment logic                              │
│     • Sends result to resultChan                             │
│                                                               │
│  4. Select statement waits for:                              │
│     • ctx.Done() → Timeout error                             │
│     • errChan    → Assignment error                          │
│     • resultChan → Success result                            │
│                                                               │
│  5. Context cancelled (cleanup)                              │
│     defer cancel()                                           │
└──────────────────────────────────────────────────────────────┘
```

## Deployment Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    DOCKER CONTAINER                           │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  Alpine Linux (minimal)                                 │  │
│  │  ┌──────────────────────────────────────────────────┐  │  │
│  │  │  task-assignment-engine binary                    │  │  │
│  │  │  • Compiled with CGO_ENABLED=0                    │  │  │
│  │  │  • Static binary                                  │  │  │
│  │  │  • Port 8080 exposed                              │  │  │
│  │  └──────────────────────────────────────────────────┘  │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
                             │
                             │ Port mapping 8080:8080
                             ▼
┌──────────────────────────────────────────────────────────────┐
│                        HOST MACHINE                           │
│  • Can be deployed to any Docker-compatible platform         │
│  • Kubernetes, Docker Swarm, cloud platforms, etc.           │
└──────────────────────────────────────────────────────────────┘
```

## API Request/Response Flow

```
Client Request → Gin Router → Handler → Business Logic → Store
     │                                                      │
     │                                                      │
     └───────────── Response ← JSON ← Result ← Data ←──────┘
```

## Key Design Patterns Used

1. **Repository Pattern** - Store abstracts data access
2. **Worker Pool Pattern** - Controlled concurrency
3. **Singleton Pattern** - Single store instance
4. **Strategy Pattern** - Pluggable distance calculation
5. **Factory Pattern** - Creating employees and tasks with UUIDs

## Technology Stack Visualization

```
┌─────────────────────────────────────────────────────────┐
│  Application Layer:    Gin Web Framework                │
├─────────────────────────────────────────────────────────┤
│  Business Logic:       Custom Go code                   │
├─────────────────────────────────────────────────────────┤
│  Data Layer:           In-Memory Maps (sync.RWMutex)    │
├─────────────────────────────────────────────────────────┤
│  Concurrency:          Goroutines + Channels            │
├─────────────────────────────────────────────────────────┤
│  Context Management:   context.Context                  │
├─────────────────────────────────────────────────────────┤
│  Testing:              Go testing package               │
├─────────────────────────────────────────────────────────┤
│  Containerization:     Docker + Docker Compose          │
└─────────────────────────────────────────────────────────┘
```

## Legend

```
┌─────┐
│     │  = Component/Module
└─────┘

   │
   ▼     = Data flow direction

  ═══    = Layer boundary

  ←→     = Bidirectional communication

  • •    = Items in a list
```

---

This architecture provides:
- ✅ Separation of concerns
- ✅ Thread safety
- ✅ Scalable concurrency
- ✅ Clean error handling
- ✅ Easy to test and maintain
