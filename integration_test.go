package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"task-management-api/internal/dto"
	"task-management-api/internal/handler"
	"task-management-api/internal/repository"
	"task-management-api/internal/service"
)

// buildServer wires the full stack (no mocks) and returns a test HTTP server.
func buildServer() http.Handler {
	repo := repository.NewInMemoryTaskRepository()
	svc := service.NewTaskService(repo)
	h := handler.NewTaskHandler(svc)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func doReq(t *testing.T, srv http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	return rr
}

func decode[T any](t *testing.T, rr *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&v))
	return v
}

// ── Full CRUD flow ────────────────────────────────────────────────────────────

func TestIntegration_CRUD_Flow(t *testing.T) {
	srv := buildServer()

	// 1. Create
	createReq := dto.CreateTaskRequest{
		Title:       "Integration Task",
		Description: "A full-flow test",
		Status:      "PENDING",
		DueDate:     time.Now().Add(48 * time.Hour).Format("2006-01-02"),
	}
	rr := doReq(t, srv, http.MethodPost, "/tasks", createReq)
	require.Equal(t, http.StatusCreated, rr.Code)
	created := decode[dto.TaskResponse](t, rr)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "Integration Task", created.Title)
	assert.Equal(t, "PENDING", created.Status)

	id := created.ID

	// 2. Get
	rr = doReq(t, srv, http.MethodGet, "/tasks/"+id, nil)
	require.Equal(t, http.StatusOK, rr.Code)
	fetched := decode[dto.TaskResponse](t, rr)
	assert.Equal(t, id, fetched.ID)
	assert.Equal(t, "Integration Task", fetched.Title)

	// 3. Update (partial)
	newTitle := "Updated Integration Task"
	newStatus := "IN_PROGRESS"
	rr = doReq(t, srv, http.MethodPut, "/tasks/"+id, dto.UpdateTaskRequest{
		Title:  &newTitle,
		Status: &newStatus,
	})
	require.Equal(t, http.StatusOK, rr.Code)
	updated := decode[dto.TaskResponse](t, rr)
	assert.Equal(t, "Updated Integration Task", updated.Title)
	assert.Equal(t, "IN_PROGRESS", updated.Status)

	// 4. List — task must appear
	rr = doReq(t, srv, http.MethodGet, "/tasks", nil)
	require.Equal(t, http.StatusOK, rr.Code)
	list := decode[dto.TaskListResponse](t, rr)
	assert.Equal(t, 1, list.Total)
	assert.Len(t, list.Tasks, 1)

	// 5. Delete
	rr = doReq(t, srv, http.MethodDelete, "/tasks/"+id, nil)
	assert.Equal(t, http.StatusNoContent, rr.Code)

	// 6. Get after delete → 404
	rr = doReq(t, srv, http.MethodGet, "/tasks/"+id, nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ── List: sorting + filtering + pagination ────────────────────────────────────

func TestIntegration_List_SortFilterPaginate(t *testing.T) {
	srv := buildServer()

	statuses := []string{"PENDING", "IN_PROGRESS", "DONE"}
	// Create 6 tasks with staggered due dates and cycling statuses.
	for i := 1; i <= 6; i++ {
		req := dto.CreateTaskRequest{
			Title:   fmt.Sprintf("Task %d", i),
			Status:  statuses[(i-1)%3],
			DueDate: time.Now().Add(time.Duration(i*24) * time.Hour).Format("2006-01-02"),
		}
		rr := doReq(t, srv, http.MethodPost, "/tasks", req)
		require.Equal(t, http.StatusCreated, rr.Code)
	}

	// All tasks, sorted by due_date ascending
	rr := doReq(t, srv, http.MethodGet, "/tasks?page=1&limit=10", nil)
	require.Equal(t, http.StatusOK, rr.Code)
	all := decode[dto.TaskListResponse](t, rr)
	assert.Equal(t, 6, all.Total)
	// Verify ascending due_date order
	for i := 1; i < len(all.Tasks); i++ {
		d1, _ := time.Parse("2006-01-02", all.Tasks[i-1].DueDate)
		d2, _ := time.Parse("2006-01-02", all.Tasks[i].DueDate)
		assert.True(t, !d1.After(d2), "tasks not sorted by due_date")
	}

	// Filter by DONE (tasks 3 and 6)
	rr = doReq(t, srv, http.MethodGet, "/tasks?status=DONE", nil)
	require.Equal(t, http.StatusOK, rr.Code)
	filtered := decode[dto.TaskListResponse](t, rr)
	assert.Equal(t, 2, filtered.Total)
	for _, task := range filtered.Tasks {
		assert.Equal(t, "DONE", task.Status)
	}

	// Pagination: page 1 limit 2
	rr = doReq(t, srv, http.MethodGet, "/tasks?page=1&limit=2", nil)
	require.Equal(t, http.StatusOK, rr.Code)
	page1 := decode[dto.TaskListResponse](t, rr)
	assert.Equal(t, 6, page1.Total)
	assert.Len(t, page1.Tasks, 2)
	assert.Equal(t, 3, page1.TotalPages)

	// Pagination: page 3 limit 2 (last page)
	rr = doReq(t, srv, http.MethodGet, "/tasks?page=3&limit=2", nil)
	require.Equal(t, http.StatusOK, rr.Code)
	page3 := decode[dto.TaskListResponse](t, rr)
	assert.Len(t, page3.Tasks, 2)
}

// ── Validation edge cases ─────────────────────────────────────────────────────

func TestIntegration_Validation_MissingTitle(t *testing.T) {
	srv := buildServer()
	rr := doReq(t, srv, http.MethodPost, "/tasks", dto.CreateTaskRequest{
		DueDate: time.Now().Add(48 * time.Hour).Format("2006-01-02"),
	})
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	errResp := decode[dto.ErrorResponse](t, rr)
	assert.Contains(t, errResp.Error, "title")
}

func TestIntegration_Validation_PastDueDate(t *testing.T) {
	srv := buildServer()
	rr := doReq(t, srv, http.MethodPost, "/tasks", dto.CreateTaskRequest{
		Title:   "Past Task",
		DueDate: time.Now().Add(-24 * time.Hour).Format("2006-01-02"),
	})
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	errResp := decode[dto.ErrorResponse](t, rr)
	assert.Contains(t, errResp.Error, "future")
}

func TestIntegration_Validation_InvalidStatus(t *testing.T) {
	srv := buildServer()
	rr := doReq(t, srv, http.MethodPost, "/tasks", dto.CreateTaskRequest{
		Title:   "Task",
		DueDate: time.Now().Add(48 * time.Hour).Format("2006-01-02"),
		Status:  "UNKNOWN",
	})
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestIntegration_GetNonExistent_404(t *testing.T) {
	srv := buildServer()
	rr := doReq(t, srv, http.MethodGet, "/tasks/does-not-exist", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestIntegration_DeleteNonExistent_404(t *testing.T) {
	srv := buildServer()
	rr := doReq(t, srv, http.MethodDelete, "/tasks/does-not-exist", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestIntegration_UpdateNonExistent_404(t *testing.T) {
	srv := buildServer()
	title := "x"
	rr := doReq(t, srv, http.MethodPut, "/tasks/does-not-exist", dto.UpdateTaskRequest{Title: &title})
	assert.Equal(t, http.StatusNotFound, rr.Code)
}
