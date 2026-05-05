# Task Management REST API

A production-quality Task Management REST API built in **Go** following **Domain-Driven Design (DDD)**, **Test-Driven Development (TDD)**, and **Clean Architecture** principles.

---

## Tech Stack

| Concern       | Choice                        |
|---------------|-------------------------------|
| Language      | Go 1.21+                      |
| HTTP          | `net/http` (stdlib)           |
| Storage       | In-memory (thread-safe)       |
| Testing       | `testing` + `testify`         |
| UUID          | `github.com/google/uuid`      |

---

## Project Structure

```
.
├── cmd/server/main.go                          # Entry point
├── internal/
│   ├── domain/
│   │   ├── task.go                             # Task aggregate + Validate()
│   │   └── status.go                           # Status enum + ParseStatus()
│   ├── repository/
│   │   ├── task_repository.go                  # Repository interface
│   │   ├── in_memory_task_repository.go        # Thread-safe in-memory impl
│   │   └── in_memory_task_repository_test.go   # Repository unit tests
│   ├── service/
│   │   ├── task_service.go                     # Use-case layer
│   │   ├── mock_task_repository.go             # testify mock for repo
│   │   └── task_service_test.go                # Service unit tests
│   ├── handler/
│   │   ├── task_handler.go                     # HTTP handler
│   │   ├── mock_task_service.go                # testify mock for service
│   │   └── task_handler_test.go                # Handler unit tests
│   ├── dto/
│   │   ├── task_request.go                     # CreateTaskRequest, UpdateTaskRequest
│   │   └── task_response.go                    # TaskResponse, TaskListResponse, ErrorResponse
│   ├── errors/
│   │   └── errors.go                           # Sentinel errors
│   └── utils/
│       └── validator.go                        # ParseFutureDate helper
└── integration_test.go                         # Full-stack integration tests
```

---

## How to Run

```bash
# Clone / navigate to the project
cd india-satcom-coding-challange

# Install dependencies
go mod tidy

# Run the server (listens on :8083)
go run cmd/server/main.go
```

---

## How to Test

```bash
# Run ALL tests (unit + integration) with verbose output
go test ./... -v

# Run only a specific package
go test ./internal/service/... -v
go test ./internal/handler/... -v
go test ./internal/repository/... -v

# Run only integration tests
go test -v -run TestIntegration

# Run with race detector
go test -race ./...
```

### Test Coverage Summary

| Layer       | Tests                                                   |
|-------------|--------------------------------------------------------|
| Repository  | Save, Get, Update, Delete, GetAll, copy isolation      |
| Service     | Create (happy + all validation paths), Get, Update (partial, not-found, bad status), Delete, List (sort, filter, paginate) |
| Handler     | 201 Create, 400 validation, 404 not-found, 204 delete, 200 list with pagination & filter, 405 method-not-allowed |
| Integration | Full CRUD flow, sort+filter+paginate, all validation edge cases |

---

## API Reference

### Domain Model

```
Task {
  id          string   // UUID, auto-generated
  title       string   // required
  description string   // optional
  status      enum     // PENDING | IN_PROGRESS | DONE (default: PENDING)
  due_date    string   // YYYY-MM-DD, must be in the future
}
```

### Endpoints

#### `POST /tasks` — Create a task

```bash
curl -s -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Design API",
    "description": "Draw the OpenAPI spec",
    "status": "IN_PROGRESS",
    "due_date": "2027-01-15"
  }' | jq
```

**Response 201:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Design API",
  "description": "Draw the OpenAPI spec",
  "status": "IN_PROGRESS",
  "due_date": "2027-01-15"
}
```

---

#### `GET /tasks/{id}` — Get a task

```bash
curl -s http://localhost:8080/tasks/550e8400-e29b-41d4-a716-446655440000 | jq
```

**Response 200:** Task object (see above)  
**Response 404:** `{"error": "task not found"}`

---

#### `PUT /tasks/{id}` — Update a task (partial)

```bash
curl -s -X PUT http://localhost:8080/tasks/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{"status": "DONE"}' | jq
```

**Response 200:** Updated task object  
**Response 400:** `{"error": "..."}`  
**Response 404:** `{"error": "task not found"}`

---

#### `DELETE /tasks/{id}` — Delete a task

```bash
curl -s -o /dev/null -w "%{http_code}" \
  -X DELETE http://localhost:8080/tasks/550e8400-e29b-41d4-a716-446655440000
# → 204
```

**Response 204:** No content  
**Response 404:** `{"error": "task not found"}`

---

#### `GET /tasks` — List all tasks

```bash
# All tasks sorted by due_date
curl -s http://localhost:8080/tasks | jq

# With pagination
curl -s "http://localhost:8080/tasks?page=1&limit=5" | jq

# Filter by status
curl -s "http://localhost:8080/tasks?status=PENDING" | jq

# Combined
curl -s "http://localhost:8080/tasks?page=2&limit=5&status=IN_PROGRESS" | jq
```

**Query parameters:**

| Param    | Default | Description                            |
|----------|---------|----------------------------------------|
| `page`   | `1`     | Page number (1-indexed)               |
| `limit`  | `10`    | Results per page                       |
| `status` | —       | Filter: `PENDING`, `IN_PROGRESS`, `DONE` |

**Response 200:**
```json
{
  "tasks": [...],
  "total": 12,
  "page": 2,
  "limit": 5,
  "total_pages": 3
}
```

---

## Validation Rules

| Field      | Rule                                                     |
|------------|----------------------------------------------------------|
| `title`    | Required, non-empty                                      |
| `due_date` | Required, format `YYYY-MM-DD`, must be a future date     |
| `status`   | Must be `PENDING`, `IN_PROGRESS`, or `DONE`              |

---

## Error Format

All errors follow a consistent envelope:

```json
{ "error": "descriptive message" }
```

---

## Design Decisions

- **DDD layers**: domain → repository (interface) → service → handler. No layer reaches backwards.
- **Dependency injection**: every struct receives its dependencies via constructor (`NewXxx()`). No `init()` or package-level singletons.
- **In-memory store**: protected by `sync.RWMutex`; stored copies prevent external mutation.
- **Partial update**: `UpdateTaskRequest` fields are pointers; `nil` means "leave unchanged".
- **Pagination**: applied after filtering and sorting so `total` always reflects the filtered count.
