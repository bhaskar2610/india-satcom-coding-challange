package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"

	"task-management-api/internal/dto"
	apperrors "task-management-api/internal/errors"
	"task-management-api/internal/service"
)

// TaskHandler wires HTTP routes to service use-cases.
type TaskHandler struct {
	svc service.TaskService
}

// NewTaskHandler constructs a handler with its dependency injected.
func NewTaskHandler(svc service.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

// RegisterRoutes attaches all task endpoints to the given mux.
func (h *TaskHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/tasks", h.handleTasks)
	mux.HandleFunc("/tasks/", h.handleTaskByID)
}

// handleTasks dispatches GET /tasks and POST /tasks.
func (h *TaskHandler) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listTasks(w, r)
	case http.MethodPost:
		h.createTask(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleTaskByID dispatches GET/PUT/DELETE /tasks/{id}.
func (h *TaskHandler) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "task id is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getTask(w, r, id)
	case http.MethodPut:
		h.updateTask(w, r, id)
	case http.MethodDelete:
		h.deleteTask(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// createTask handles POST /tasks.
func (h *TaskHandler) createTask(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	task, err := h.svc.CreateTask(req)
	if err != nil {
		if service.IsValidationError(err) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, apperrors.ErrInternalServer.Error())
		return
	}

	writeJSON(w, http.StatusCreated, dto.FromDomain(task))
}

// getTask handles GET /tasks/{id}.
func (h *TaskHandler) getTask(w http.ResponseWriter, r *http.Request, id string) {
	task, err := h.svc.GetTaskByID(id)
	if err != nil {
		if service.IsNotFound(err) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, apperrors.ErrInternalServer.Error())
		return
	}
	writeJSON(w, http.StatusOK, dto.FromDomain(task))
}

// updateTask handles PUT /tasks/{id}.
func (h *TaskHandler) updateTask(w http.ResponseWriter, r *http.Request, id string) {
	var req dto.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	task, err := h.svc.UpdateTask(id, req)
	if err != nil {
		if service.IsNotFound(err) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		if service.IsValidationError(err) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, apperrors.ErrInternalServer.Error())
		return
	}

	writeJSON(w, http.StatusOK, dto.FromDomain(task))
}

// deleteTask handles DELETE /tasks/{id}.
func (h *TaskHandler) deleteTask(w http.ResponseWriter, r *http.Request, id string) {
	err := h.svc.DeleteTask(id)
	if err != nil {
		if service.IsNotFound(err) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, apperrors.ErrInternalServer.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// listTasks handles GET /tasks with optional ?page=, ?limit=, ?status= params.
func (h *TaskHandler) listTasks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 {
		limit = 10
	}
	statusFilter := q.Get("status")

	opts := service.ListOptions{
		Page:   page,
		Limit:  limit,
		Status: statusFilter,
	}

	tasks, total, err := h.svc.ListTasks(opts)
	if err != nil {
		if service.IsValidationError(err) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, apperrors.ErrInternalServer.Error())
		return
	}

	responses := make([]dto.TaskResponse, len(tasks))
	for i, t := range tasks {
		responses[i] = dto.FromDomain(t)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	writeJSON(w, http.StatusOK, dto.TaskListResponse{
		Tasks:      responses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, dto.ErrorResponse{Error: message})
}
