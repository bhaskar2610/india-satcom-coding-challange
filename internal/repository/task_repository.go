package repository

import "github.com/bhaskar2610/india-satcom-coding-challange/internal/domain"

// TaskRepository defines the persistence contract for Task aggregates.
// Any concrete storage backend (in-memory, SQL, NoSQL) must implement this
// interface, keeping the domain and service layers storage-agnostic.
type TaskRepository interface {
	Save(task *domain.Task) error
	GetByID(id string) (*domain.Task, error)
	Update(task *domain.Task) error
	Delete(id string) error
	GetAll() ([]*domain.Task, error)
}
