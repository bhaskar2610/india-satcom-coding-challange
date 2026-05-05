package repository

import (
	"sync"

	apperrors "task-management-api/internal/errors"
	"task-management-api/internal/domain"
)

// InMemoryTaskRepository is a thread-safe, in-process implementation of
// TaskRepository backed by a plain Go map.
type InMemoryTaskRepository struct {
	mu    sync.RWMutex
	store map[string]*domain.Task
}

// NewInMemoryTaskRepository returns an initialised repository ready for use.
func NewInMemoryTaskRepository() *InMemoryTaskRepository {
	return &InMemoryTaskRepository{
		store: make(map[string]*domain.Task),
	}
}

// Save persists a new task. It stores a copy to prevent external mutation.
func (r *InMemoryTaskRepository) Save(task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := *task
	r.store[task.ID] = &copy
	return nil
}

// GetByID retrieves a task by its ID.
func (r *InMemoryTaskRepository) GetByID(id string) (*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.store[id]
	if !ok {
		return nil, apperrors.ErrTaskNotFound
	}
	copy := *t
	return &copy, nil
}

// Update replaces the stored task with the provided value.
func (r *InMemoryTaskRepository) Update(task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[task.ID]; !ok {
		return apperrors.ErrTaskNotFound
	}
	copy := *task
	r.store[task.ID] = &copy
	return nil
}

// Delete removes the task identified by id.
func (r *InMemoryTaskRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[id]; !ok {
		return apperrors.ErrTaskNotFound
	}
	delete(r.store, id)
	return nil
}

// GetAll returns a snapshot of every task in the store.
func (r *InMemoryTaskRepository) GetAll() ([]*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tasks := make([]*domain.Task, 0, len(r.store))
	for _, t := range r.store {
		copy := *t
		tasks = append(tasks, &copy)
	}
	return tasks, nil
}
