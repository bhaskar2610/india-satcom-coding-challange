package domain

import (
	"errors"
	"time"
)

// Task is the core domain entity.
type Task struct {
	ID          string
	Title       string
	Description string
	Status      Status
	DueDate     time.Time
}

// Validate enforces domain invariants on a Task.
func (t *Task) Validate() error {
	if t.Title == "" {
		return errors.New("title is required")
	}
	if t.DueDate.IsZero() {
		return errors.New("due_date is required")
	}
	if !t.DueDate.After(time.Now()) {
		return errors.New("due_date must be in the future")
	}
	if !t.Status.IsValid() {
		return errors.New("status must be one of PENDING, IN_PROGRESS, DONE")
	}
	return nil
}
