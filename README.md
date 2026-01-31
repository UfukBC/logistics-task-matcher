# Task Assignment Engine API

A high-performance, production-ready Task Assignment Engine API for logistics companies, built with Go. This system intelligently assigns tasks to the nearest available employees based on location and skills.

## 🏗️ Architecture Overview

### Design Decisions

This project follows **Clean Architecture** principles with a pragmatic approach suitable for rapid development:

1. **In-Memory Storage**: Uses thread-safe in-memory maps with `sync.RWMutex` for fast data access without database overhead
2. **Concurrency Model**: Implements worker pool pattern with goroutines and channels for asynchronous task assignment
3. **Distance Calculation**: Uses the Haversine formula for accurate geographical distance calculations
4. **Context-Based Timeout Management**: Leverages `context.Context` for proper timeout handling and graceful cancellation
5. **Custom Error Types**: Implements structured error handling for better debugging and API responses
6. **RESTful API**: Built with Gin framework for high-performance HTTP routing

### Key Components

```
task-assignment-engine/
├── models.go       # Data models, business logic, and storage layer
├── main.go         # API handlers, routing, and server setup
├── models_test.go  # Comprehensive unit tests
├── Dockerfile      # Multi-stage Docker build
└── docker-compose.yml
```

#### 1. Data Layer (`models.go`)
- **Employee & Task Models**: Structured data with validation
- **Store**: Thread-safe in-memory storage with RWMutex
- **Custom Errors**: Type-safe error handling

#### 2. Business Logic (`models.go`)
- **CalculateDistance**: Haversine formula implementation
- **TaskAssigner**: Core assignment algorithm
- **AssignmentWorkerPool**: Concurrent task processing

#### 3. API Layer (`main.go`)
- **Gin Router**: RESTful endpoints
- **Graceful Shutdown**: Signal handling for clean termination
- **CORS Support**: Cross-origin request handling

## 🚀 Features

- ✅ **Thread-Safe Operations**: No data races with proper mutex usage
- ✅ **Asynchronous Processing**: Non-blocking task assignments
- ✅ **Distance-Based Matching**: Assigns tasks to the closest available employee
- ✅ **Skill Filtering**: Matches employees based on required skills
- ✅ **Context Timeout Management**: Prevents long-running operations
- ✅ **Health Check Endpoint**: For monitoring and load balancers
- ✅ **Comprehensive Testing**: Unit tests and benchmarks included
- ✅ **Docker Support**: Easy deployment with Docker and Docker Compose

## 📋 API Endpoints

### 1. Health Check
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "time": "2026-01-31T12:00:00Z"
}
```

### 2. Create Employee
```http
POST /employees
Content-Type: application/json

{
  "name": "John Doe",
  "location": {
    "lat": 60.1699,
    "lon": 24.9384
  },
  "skills": ["delivery", "driving"]
}
```

**Response:**
```json
{
  "message": "Employee created successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Doe",
    "location": {
      "lat": 60.1699,
      "lon": 24.9384
    },
    "skills": ["delivery", "driving"],
    "is_available": true
  }
}
```

### 3. Get All Employees
```http
GET /employees
```

**Response:**
```json
{
  "message": "Retrieved 5 employees",
  "data": [...]
}
```

### 4. Create Task (Triggers Assignment)
```http
POST /tasks
Content-Type: application/json

{
  "location": {
    "lat": 60.1700,
    "lon": 24.9400
  },
  "required_skill": "delivery"
}
```

**Response:**
```json
{
  "message": "Task created and assignment initiated",
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "location": {
      "lat": 60.1700,
      "lon": 24.9400
    },
    "required_skill": "delivery",
    "status": "pending"
  }
}
```

### 5. Get All Tasks
```http
GET /tasks
```

**Response:**
```json
{
  "message": "Retrieved 10 tasks",
  "data": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440000",
      "location": {
        "lat": 60.1700,
        "lon": 24.9400
      },
      "required_skill": "delivery",
      "status": "assigned",
      "assigned_employee_id": "550e8400-e29b-41d4-a716-446655440000"
    }
  ]
}
```

### 6. Get Task by ID
```http
GET /tasks/:id
```

**Response:**
```json
{
  "message": "Task retrieved successfully",
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "location": {
      "lat": 60.1700,
      "lon": 24.9400
    },
    "required_skill": "delivery",
    "status": "assigned",
    "assigned_employee_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

## 🔧 Installation & Setup

### Prerequisites
- Go 1.21 or higher
- Docker & Docker Compose (optional)

### Option 1: Run with Go

1. **Clone and navigate to the project:**
```bash
git clone <repository-url>
cd task-assignment-engine
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Run the application:**
```bash
go run .
```

The server will start on `http://localhost:8080`

### Option 2: Run with Docker

1. **Build and run with Docker Compose:**
```bash
docker-compose up --build
```

2. **Run detached:**
```bash
docker-compose up -d
```

3. **View logs:**
```bash
docker-compose logs -f
```

4. **Stop the service:**
```bash
docker-compose down
```

### Option 3: Build Docker Image Manually

```bash
# Build the image
docker build -t task-assignment-engine .

# Run the container
docker run -p 8080:8080 task-assignment-engine
```

## 🧪 Testing

### Run All Tests
```bash
go test -v
```

### Run Tests with Coverage
```bash
go test -v -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run Benchmarks
```bash
go test -bench=. -benchmem
```

### Test Coverage Includes:
- ✅ Haversine distance calculation accuracy
- ✅ Skill matching logic
- ✅ Thread-safe store operations
- ✅ Task assignment algorithm
- ✅ Concurrent operations (race detection)
- ✅ Context timeout handling
- ✅ Error handling scenarios

### Run Race Detector
```bash
go test -race
```

## 📊 Performance Characteristics

### Concurrency Model
- **Worker Pool**: 5 concurrent workers by default
- **Channel Buffer**: 100 task queue capacity
- **Assignment Timeout**: 30 seconds per task

### Time Complexity
- **Distance Calculation**: O(1) - Constant time Haversine formula
- **Employee Search**: O(n) - Linear scan through available employees
- **Task Assignment**: O(n) - Where n is the number of eligible employees

### Optimization Opportunities
For production at scale, consider:
1. **Spatial Indexing**: Use R-tree or geohashing for faster location queries
2. **Caching**: Cache distance calculations for frequently queried locations
3. **Database**: Replace in-memory store with PostgreSQL + PostGIS for persistence
4. **Message Queue**: Use RabbitMQ/Kafka for distributed task processing
5. **Load Balancing**: Deploy multiple instances behind a load balancer

## 🔒 Production Considerations

### Security
- Add authentication/authorization (JWT, OAuth2)
- Implement rate limiting
- Add input validation and sanitization
- Use HTTPS/TLS in production

### Monitoring
- Add structured logging (logrus, zap)
- Implement metrics (Prometheus)
- Add distributed tracing (OpenTelemetry)
- Set up health checks and readiness probes

### Scalability
- Deploy multiple instances with load balancer
- Use Redis for shared state in distributed setup
- Implement horizontal pod autoscaling (HPA)
- Add database connection pooling

### Reliability
- Implement circuit breakers
- Add retry logic with exponential backoff
- Use dead letter queues for failed assignments
- Implement idempotency keys

## 🎯 Assignment Algorithm

### How It Works

1. **Task Creation**: When a task is created via POST `/tasks`, it's added to the store with `pending` status
2. **Async Processing**: The task is submitted to the worker pool for asynchronous processing
3. **Filtering**: Workers filter employees by:
   - Availability (`is_available = true`)
   - Required skill match
4. **Distance Calculation**: For each eligible employee, calculate distance using Haversine formula
5. **Selection**: Assign task to the closest employee
6. **State Update**:
   - Task status → `assigned`
   - Employee availability → `false`
   - Task's `assigned_employee_id` set

### Haversine Formula

The distance between two points on a sphere is calculated as:

```
a = sin²(Δφ/2) + cos φ₁ × cos φ₂ × sin²(Δλ/2)
c = 2 × atan2(√a, √(1−a))
d = R × c
```

Where:
- φ is latitude
- λ is longitude
- R is Earth's radius (6,371 km)
- All angles in radians

## 📝 Example Usage

### Complete Workflow

```bash
# 1. Check health
curl http://localhost:8080/health

# 2. Create employees
curl -X POST http://localhost:8080/employees \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice",
    "location": {"lat": 60.1699, "lon": 24.9384},
    "skills": ["delivery", "driving"]
  }'

curl -X POST http://localhost:8080/employees \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Bob",
    "location": {"lat": 60.2055, "lon": 24.6559},
    "skills": ["delivery"]
  }'

# 3. Create a task (triggers assignment)
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "location": {"lat": 60.1700, "lon": 24.9400},
    "required_skill": "delivery"
  }'

# 4. Check task status
curl http://localhost:8080/tasks

# 5. Check employees (see availability status)
curl http://localhost:8080/employees
```

## 🛠️ Troubleshooting

### Port Already in Use
```bash
# Change the port using environment variable
PORT=8081 go run .
```

### Dependencies Not Found
```bash
go mod tidy
go mod download
```

### Docker Build Fails
```bash
# Clean Docker cache and rebuild
docker-compose down
docker system prune -a
docker-compose up --build
```

## 📚 Technology Stack

- **Language**: Go 1.21
- **Web Framework**: Gin
- **Concurrency**: Goroutines & Channels
- **Testing**: Go's built-in testing package
- **Containerization**: Docker & Docker Compose
- **Architecture**: Clean Architecture (simplified)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License.

## 🎓 Learning Resources

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Clean Architecture by Uncle Bob](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Haversine Formula](https://en.wikipedia.org/wiki/Haversine_formula)

---

**Note**: This is a production-ready implementation with proper error handling, testing, and documentation. The in-memory store is intentionally used for rapid development, but the architecture supports easy migration to a persistent database.
