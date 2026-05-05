package service_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bhaskar2610/india-satcom-coding-challange/internal/domain"
	apperrors "github.com/bhaskar2610/india-satcom-coding-challange/internal/errors"
	"github.com/bhaskar2610/india-satcom-coding-challange/internal/dto"
	"github.com/bhaskar2610/india-satcom-coding-challange/internal/service"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func tomorrow() string {
	return time.Now().Add(48 * time.Hour).Format("2006-01-02")
}

func yesterday() string {
	return time.Now().Add(-24 * time.Hour).Format("2006-01-02")
}

func newMockRepo() *service.MockTaskRepository {
	return new(service.MockTaskRepository)
}

// ── CreateTask ────────────────────────────────────────────────────────────────

func TestCreateTask_Success(t *testing.T) {
	repo := newMockRepo()
	repo.On("Save", mock.AnythingOfType("*domain.Task")).Return(nil)

	svc := service.NewTaskService(repo)
	req := dto.CreateTaskRequest{
		Title:   "My Task",
		Status:  "PENDING",
		DueDate: tomorrow(),
	}

	task, err := svc.CreateTask(req)
	require.NoError(t, err)
	assert.NotEmpty(t, task.ID)
	assert.Equal(t, "My Task", task.Title)
	assert.Equal(t, domain.StatusPending, task.Status)
	repo.AssertExpectations(t)
}

func TestCreateTask_DefaultsStatusToPending(t *testing.T) {
	repo := newMockRepo()
	repo.On("Save", mock.AnythingOfType("*domain.Task")).Return(nil)

	svc := service.NewTaskService(repo)
	req := dto.CreateTaskRequest{
		Title:   "No Status",
		DueDate: tomorrow(),
	}

	task, err := svc.CreateTask(req)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusPending, task.Status)
}

func TestCreateTask_MissingTitle(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewTaskService(repo)

	_, err := svc.CreateTask(dto.CreateTaskRequest{DueDate: tomorrow()})
	require.Error(t, err)
	assert.True(t, service.IsValidationError(err))
	assert.Contains(t, err.Error(), "title")
}

func TestCreateTask_MissingDueDate(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewTaskService(repo)

	_, err := svc.CreateTask(dto.CreateTaskRequest{Title: "T"})
	require.Error(t, err)
	assert.True(t, service.IsValidationError(err))
}

func TestCreateTask_PastDueDate(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewTaskService(repo)

	_, err := svc.CreateTask(dto.CreateTaskRequest{Title: "T", DueDate: yesterday()})
	require.Error(t, err)
	assert.True(t, service.IsValidationError(err))
	assert.Contains(t, err.Error(), "future")
}

func TestCreateTask_InvalidStatus(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewTaskService(repo)

	_, err := svc.CreateTask(dto.CreateTaskRequest{
		Title:   "T",
		DueDate: tomorrow(),
		Status:  "INVALID",
	})
	require.Error(t, err)
	assert.True(t, service.IsValidationError(err))
}

func TestCreateTask_InvalidDueDateFormat(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewTaskService(repo)

	_, err := svc.CreateTask(dto.CreateTaskRequest{
		Title:   "T",
		DueDate: "31-12-2099", // wrong format
	})
	require.Error(t, err)
	assert.True(t, service.IsValidationError(err))
}

// ── GetTaskByID ───────────────────────────────────────────────────────────────

func TestGetTaskByID_Success(t *testing.T) {
	repo := newMockRepo()
	stored := &domain.Task{ID: "abc", Title: "Hello", Status: domain.StatusPending, DueDate: time.Now().Add(24 * time.Hour)}
	repo.On("GetByID", "abc").Return(stored, nil)

	svc := service.NewTaskService(repo)
	got, err := svc.GetTaskByID("abc")
	require.NoError(t, err)
	assert.Equal(t, "abc", got.ID)
	repo.AssertExpectations(t)
}

func TestGetTaskByID_NotFound(t *testing.T) {
	repo := newMockRepo()
	repo.On("GetByID", "missing").Return(nil, apperrors.ErrTaskNotFound)

	svc := service.NewTaskService(repo)
	_, err := svc.GetTaskByID("missing")
	assert.ErrorIs(t, err, apperrors.ErrTaskNotFound)
	assert.True(t, service.IsNotFound(err))
}

// ── UpdateTask ────────────────────────────────────────────────────────────────

func TestUpdateTask_PartialUpdate(t *testing.T) {
	repo := newMockRepo()
	stored := &domain.Task{ID: "u1", Title: "Old", Status: domain.StatusPending, DueDate: time.Now().Add(48 * time.Hour)}
	repo.On("GetByID", "u1").Return(stored, nil)
	repo.On("Update", mock.AnythingOfType("*domain.Task")).Return(nil)

	svc := service.NewTaskService(repo)
	newTitle := "New Title"
	got, err := svc.UpdateTask("u1", dto.UpdateTaskRequest{Title: &newTitle})
	require.NoError(t, err)
	assert.Equal(t, "New Title", got.Title)
	repo.AssertExpectations(t)
}

func TestUpdateTask_InvalidStatus(t *testing.T) {
	repo := newMockRepo()
	stored := &domain.Task{ID: "u2", Title: "T", Status: domain.StatusPending, DueDate: time.Now().Add(48 * time.Hour)}
	repo.On("GetByID", "u2").Return(stored, nil)

	svc := service.NewTaskService(repo)
	bad := "BAD_STATUS"
	_, err := svc.UpdateTask("u2", dto.UpdateTaskRequest{Status: &bad})
	require.Error(t, err)
	assert.True(t, service.IsValidationError(err))
}

func TestUpdateTask_NotFound(t *testing.T) {
	repo := newMockRepo()
	repo.On("GetByID", "ghost").Return(nil, apperrors.ErrTaskNotFound)

	svc := service.NewTaskService(repo)
	title := "X"
	_, err := svc.UpdateTask("ghost", dto.UpdateTaskRequest{Title: &title})
	assert.ErrorIs(t, err, apperrors.ErrTaskNotFound)
}

func TestUpdateTask_EmptyTitleRejected(t *testing.T) {
	repo := newMockRepo()
	stored := &domain.Task{ID: "u3", Title: "T", Status: domain.StatusPending, DueDate: time.Now().Add(48 * time.Hour)}
	repo.On("GetByID", "u3").Return(stored, nil)

	svc := service.NewTaskService(repo)
	empty := ""
	_, err := svc.UpdateTask("u3", dto.UpdateTaskRequest{Title: &empty})
	require.Error(t, err)
	assert.True(t, service.IsValidationError(err))
}

// ── DeleteTask ────────────────────────────────────────────────────────────────

func TestDeleteTask_Success(t *testing.T) {
	repo := newMockRepo()
	repo.On("Delete", "del1").Return(nil)

	svc := service.NewTaskService(repo)
	err := svc.DeleteTask("del1")
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteTask_NotFound(t *testing.T) {
	repo := newMockRepo()
	repo.On("Delete", "ghost").Return(apperrors.ErrTaskNotFound)

	svc := service.NewTaskService(repo)
	err := svc.DeleteTask("ghost")
	assert.ErrorIs(t, err, apperrors.ErrTaskNotFound)
}

// ── ListTasks ─────────────────────────────────────────────────────────────────

func TestListTasks_SortedByDueDate(t *testing.T) {
	repo := newMockRepo()
	now := time.Now()
	tasks := []*domain.Task{
		{ID: "3", Title: "C", Status: domain.StatusPending, DueDate: now.Add(72 * time.Hour)},
		{ID: "1", Title: "A", Status: domain.StatusPending, DueDate: now.Add(24 * time.Hour)},
		{ID: "2", Title: "B", Status: domain.StatusPending, DueDate: now.Add(48 * time.Hour)},
	}
	repo.On("GetAll").Return(tasks, nil)

	svc := service.NewTaskService(repo)
	got, total, err := svc.ListTasks(service.ListOptions{Page: 1, Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Equal(t, "A", got[0].Title)
	assert.Equal(t, "B", got[1].Title)
	assert.Equal(t, "C", got[2].Title)
}

func TestListTasks_FilterByStatus(t *testing.T) {
	repo := newMockRepo()
	now := time.Now()
	tasks := []*domain.Task{
		{ID: "1", Title: "A", Status: domain.StatusPending, DueDate: now.Add(24 * time.Hour)},
		{ID: "2", Title: "B", Status: domain.StatusDone, DueDate: now.Add(48 * time.Hour)},
	}
	repo.On("GetAll").Return(tasks, nil)

	svc := service.NewTaskService(repo)
	got, total, err := svc.ListTasks(service.ListOptions{Page: 1, Limit: 10, Status: "DONE"})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "B", got[0].Title)
}

func TestListTasks_Pagination(t *testing.T) {
	repo := newMockRepo()
	now := time.Now()
	tasks := []*domain.Task{
		{ID: "1", Title: "A", Status: domain.StatusPending, DueDate: now.Add(1 * time.Hour)},
		{ID: "2", Title: "B", Status: domain.StatusPending, DueDate: now.Add(2 * time.Hour)},
		{ID: "3", Title: "C", Status: domain.StatusPending, DueDate: now.Add(3 * time.Hour)},
	}
	repo.On("GetAll").Return(tasks, nil)

	svc := service.NewTaskService(repo)
	got, total, err := svc.ListTasks(service.ListOptions{Page: 2, Limit: 2})
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	require.Len(t, got, 1)
	assert.Equal(t, "C", got[0].Title)
}
