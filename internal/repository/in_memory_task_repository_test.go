package repository_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"task-management-api/internal/domain"
	apperrors "task-management-api/internal/errors"
	"task-management-api/internal/repository"
)

func newTask(id, title string) *domain.Task {
	return &domain.Task{
		ID:      id,
		Title:   title,
		Status:  domain.StatusPending,
		DueDate: time.Now().Add(24 * time.Hour),
	}
}

func TestInMemoryRepository_Save_And_GetByID(t *testing.T) {
	repo := repository.NewInMemoryTaskRepository()
	task := newTask("id-1", "Test Task")

	require.NoError(t, repo.Save(task))

	got, err := repo.GetByID("id-1")
	require.NoError(t, err)
	assert.Equal(t, task.ID, got.ID)
	assert.Equal(t, task.Title, got.Title)
}

func TestInMemoryRepository_GetByID_NotFound(t *testing.T) {
	repo := repository.NewInMemoryTaskRepository()

	_, err := repo.GetByID("nonexistent")
	assert.ErrorIs(t, err, apperrors.ErrTaskNotFound)
}

func TestInMemoryRepository_Update(t *testing.T) {
	repo := repository.NewInMemoryTaskRepository()
	task := newTask("id-2", "Original Title")
	require.NoError(t, repo.Save(task))

	task.Title = "Updated Title"
	require.NoError(t, repo.Update(task))

	got, err := repo.GetByID("id-2")
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", got.Title)
}

func TestInMemoryRepository_Update_NotFound(t *testing.T) {
	repo := repository.NewInMemoryTaskRepository()
	task := newTask("ghost", "Ghost Task")

	err := repo.Update(task)
	assert.ErrorIs(t, err, apperrors.ErrTaskNotFound)
}

func TestInMemoryRepository_Delete(t *testing.T) {
	repo := repository.NewInMemoryTaskRepository()
	task := newTask("id-3", "To Delete")
	require.NoError(t, repo.Save(task))

	require.NoError(t, repo.Delete("id-3"))

	_, err := repo.GetByID("id-3")
	assert.ErrorIs(t, err, apperrors.ErrTaskNotFound)
}

func TestInMemoryRepository_Delete_NotFound(t *testing.T) {
	repo := repository.NewInMemoryTaskRepository()

	err := repo.Delete("nonexistent")
	assert.ErrorIs(t, err, apperrors.ErrTaskNotFound)
}

func TestInMemoryRepository_GetAll(t *testing.T) {
	repo := repository.NewInMemoryTaskRepository()

	// Empty store
	tasks, err := repo.GetAll()
	require.NoError(t, err)
	assert.Empty(t, tasks)

	// Populate
	require.NoError(t, repo.Save(newTask("a", "Task A")))
	require.NoError(t, repo.Save(newTask("b", "Task B")))

	tasks, err = repo.GetAll()
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestInMemoryRepository_IsolatesCopies(t *testing.T) {
	// Mutations to the returned value must not affect the stored record.
	repo := repository.NewInMemoryTaskRepository()
	task := newTask("id-iso", "Isolation Test")
	require.NoError(t, repo.Save(task))

	got, err := repo.GetByID("id-iso")
	require.NoError(t, err)
	got.Title = "Mutated Outside"

	original, err := repo.GetByID("id-iso")
	require.NoError(t, err)
	assert.Equal(t, "Isolation Test", original.Title)
}
