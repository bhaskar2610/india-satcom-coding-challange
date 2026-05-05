package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bhaskar2610/india-satcom-coding-challange/internal/domain"
	apperrors "github.com/bhaskar2610/india-satcom-coding-challange/internal/errors"
	"github.com/bhaskar2610/india-satcom-coding-challange/internal/dto"
	"github.com/bhaskar2610/india-satcom-coding-challange/internal/handler"
	"github.com/bhaskar2610/india-satcom-coding-challange/internal/service"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func setupHandler(svc service.TaskService) http.Handler {
	h := handler.NewTaskHandler(svc)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func doRequest(t *testing.T, router http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func decodeBody(t *testing.T, rr *httptest.ResponseRecorder, v any) {
	t.Helper()
	require.NoError(t, json.NewDecoder(rr.Body).Decode(v))
}

func sampleTask() *domain.Task {
	return &domain.Task{
		ID:      "task-1",
		Title:   "Test Task",
		Status:  domain.StatusPending,
		DueDate: time.Now().Add(48 * time.Hour),
	}
}

// ── POST /tasks ───────────────────────────────────────────────────────────────

func TestCreateTask_201(t *testing.T) {
	svc := new(handler.MockTaskService)
	svc.On("CreateTask", mock.AnythingOfType("dto.CreateTaskRequest")).Return(sampleTask(), nil)

	router := setupHandler(svc)
	rr := doRequest(t, router, http.MethodPost, "/tasks", dto.CreateTaskRequest{
		Title:   "Test Task",
		DueDate: time.Now().Add(48 * time.Hour).Format("2006-01-02"),
	})

	assert.Equal(t, http.StatusCreated, rr.Code)
	var resp dto.TaskResponse
	decodeBody(t, rr, &resp)
	assert.Equal(t, "task-1", resp.ID)
	svc.AssertExpectations(t)
}

func TestCreateTask_ValidationError_400(t *testing.T) {
	svc := new(handler.MockTaskService)
	svc.On("CreateTask", mock.AnythingOfType("dto.CreateTaskRequest")).
		Return(nil, &service.ValidationError{Message: "title is required"})

	router := setupHandler(svc)
	rr := doRequest(t, router, http.MethodPost, "/tasks", dto.CreateTaskRequest{})

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var errResp dto.ErrorResponse
	decodeBody(t, rr, &errResp)
	assert.NotEmpty(t, errResp.Error)
}

func TestCreateTask_InvalidJSON_400(t *testing.T) {
	svc := new(handler.MockTaskService)
	router := setupHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// ── GET /tasks/{id} ───────────────────────────────────────────────────────────

func TestGetTask_200(t *testing.T) {
	svc := new(handler.MockTaskService)
	svc.On("GetTaskByID", "task-1").Return(sampleTask(), nil)

	router := setupHandler(svc)
	rr := doRequest(t, router, http.MethodGet, "/tasks/task-1", nil)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp dto.TaskResponse
	decodeBody(t, rr, &resp)
	assert.Equal(t, "task-1", resp.ID)
}

func TestGetTask_NotFound_404(t *testing.T) {
	svc := new(handler.MockTaskService)
	svc.On("GetTaskByID", "missing").Return(nil, apperrors.ErrTaskNotFound)

	router := setupHandler(svc)
	rr := doRequest(t, router, http.MethodGet, "/tasks/missing", nil)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ── PUT /tasks/{id} ───────────────────────────────────────────────────────────

func TestUpdateTask_200(t *testing.T) {
	updated := sampleTask()
	updated.Title = "Updated"

	svc := new(handler.MockTaskService)
	svc.On("UpdateTask", "task-1", mock.AnythingOfType("dto.UpdateTaskRequest")).Return(updated, nil)

	router := setupHandler(svc)
	newTitle := "Updated"
	rr := doRequest(t, router, http.MethodPut, "/tasks/task-1", dto.UpdateTaskRequest{Title: &newTitle})

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp dto.TaskResponse
	decodeBody(t, rr, &resp)
	assert.Equal(t, "Updated", resp.Title)
}

func TestUpdateTask_NotFound_404(t *testing.T) {
	svc := new(handler.MockTaskService)
	svc.On("UpdateTask", "ghost", mock.AnythingOfType("dto.UpdateTaskRequest")).
		Return(nil, apperrors.ErrTaskNotFound)

	router := setupHandler(svc)
	title := "x"
	rr := doRequest(t, router, http.MethodPut, "/tasks/ghost", dto.UpdateTaskRequest{Title: &title})

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdateTask_ValidationError_400(t *testing.T) {
	svc := new(handler.MockTaskService)
	svc.On("UpdateTask", "task-1", mock.AnythingOfType("dto.UpdateTaskRequest")).
		Return(nil, &service.ValidationError{Message: "bad status"})

	router := setupHandler(svc)
	bad := "BAD"
	rr := doRequest(t, router, http.MethodPut, "/tasks/task-1", dto.UpdateTaskRequest{Status: &bad})

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// ── DELETE /tasks/{id} ────────────────────────────────────────────────────────

func TestDeleteTask_204(t *testing.T) {
	svc := new(handler.MockTaskService)
	svc.On("DeleteTask", "task-1").Return(nil)

	router := setupHandler(svc)
	rr := doRequest(t, router, http.MethodDelete, "/tasks/task-1", nil)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestDeleteTask_NotFound_404(t *testing.T) {
	svc := new(handler.MockTaskService)
	svc.On("DeleteTask", "ghost").Return(apperrors.ErrTaskNotFound)

	router := setupHandler(svc)
	rr := doRequest(t, router, http.MethodDelete, "/tasks/ghost", nil)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ── GET /tasks ────────────────────────────────────────────────────────────────

func TestListTasks_200(t *testing.T) {
	tasks := []*domain.Task{sampleTask()}
	svc := new(handler.MockTaskService)
	svc.On("ListTasks", service.ListOptions{Page: 1, Limit: 10, Status: ""}).
		Return(tasks, 1, nil)

	router := setupHandler(svc)
	rr := doRequest(t, router, http.MethodGet, "/tasks", nil)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp dto.TaskListResponse
	decodeBody(t, rr, &resp)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.Tasks, 1)
}

func TestListTasks_WithPaginationAndFilter(t *testing.T) {
	tasks := []*domain.Task{sampleTask()}
	svc := new(handler.MockTaskService)
	svc.On("ListTasks", service.ListOptions{Page: 2, Limit: 5, Status: "PENDING"}).
		Return(tasks, 6, nil)

	router := setupHandler(svc)
	rr := doRequest(t, router, http.MethodGet, "/tasks?page=2&limit=5&status=PENDING", nil)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp dto.TaskListResponse
	decodeBody(t, rr, &resp)
	assert.Equal(t, 6, resp.Total)
	assert.Equal(t, 2, resp.TotalPages)
}

func TestListTasks_InvalidStatus_400(t *testing.T) {
	svc := new(handler.MockTaskService)
	svc.On("ListTasks", mock.AnythingOfType("service.ListOptions")).
		Return(nil, 0, &service.ValidationError{Message: "invalid status"})

	router := setupHandler(svc)
	rr := doRequest(t, router, http.MethodGet, "/tasks?status=INVALID", nil)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// ── Method not allowed ────────────────────────────────────────────────────────

func TestMethodNotAllowed_Tasks(t *testing.T) {
	svc := new(handler.MockTaskService)
	router := setupHandler(svc)
	rr := doRequest(t, router, http.MethodPatch, "/tasks", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestMethodNotAllowed_TaskByID(t *testing.T) {
	svc := new(handler.MockTaskService)
	router := setupHandler(svc)
	rr := doRequest(t, router, http.MethodPatch, fmt.Sprintf("/tasks/%s", "id-1"), nil)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}
