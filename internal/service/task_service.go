package service

import (
	"sort"

	"github.com/google/uuid"

	"github.com/bhaskar2610/india-satcom-coding-challange/internal/domain"
	apperrors "github.com/bhaskar2610/india-satcom-coding-challange/internal/errors"
	"github.com/bhaskar2610/india-satcom-coding-challange/internal/dto"
	"github.com/bhaskar2610/india-satcom-coding-challange/internal/repository"
	"github.com/bhaskar2610/india-satcom-coding-challange/internal/utils"
)

// ListOptions carries pagination and filter parameters for GetAll.
type ListOptions struct {
	Page   int
	Limit  int
	Status string // empty string means "no filter"
}

// TaskService defines the application-level use-cases for Task management.
type TaskService interface {
	CreateTask(req dto.CreateTaskRequest) (*domain.Task, error)
	GetTaskByID(id string) (*domain.Task, error)
	UpdateTask(id string, req dto.UpdateTaskRequest) (*domain.Task, error)
	DeleteTask(id string) error
	ListTasks(opts ListOptions) ([]*domain.Task, int, error)
}

type taskService struct {
	repo repository.TaskRepository
}

// NewTaskService constructs a TaskService backed by the provided repository.
func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

// CreateTask validates the request, builds a Task aggregate, and persists it.
func (s *taskService) CreateTask(req dto.CreateTaskRequest) (*domain.Task, error) {
	if req.Title == "" {
		return nil, &ValidationError{Message: "title is required"}
	}

	dueDate, err := utils.ParseFutureDate(req.DueDate)
	if err != nil {
		return nil, &ValidationError{Message: err.Error()}
	}

	// Default status to PENDING when the caller omits it.
	status := domain.StatusPending
	if req.Status != "" {
		status, err = domain.ParseStatus(req.Status)
		if err != nil {
			return nil, &ValidationError{Message: err.Error()}
		}
	}

	task := &domain.Task{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Status:      status,
		DueDate:     dueDate,
	}

	if err := s.repo.Save(task); err != nil {
		return nil, err
	}
	return task, nil
}

// GetTaskByID returns the task for the given id, or ErrTaskNotFound.
func (s *taskService) GetTaskByID(id string) (*domain.Task, error) {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// UpdateTask applies a partial update to an existing task.
func (s *taskService) UpdateTask(id string, req dto.UpdateTaskRequest) (*domain.Task, error) {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		if *req.Title == "" {
			return nil, &ValidationError{Message: "title cannot be empty"}
		}
		task.Title = *req.Title
	}

	if req.Description != nil {
		task.Description = *req.Description
	}

	if req.Status != nil {
		newStatus, parseErr := domain.ParseStatus(*req.Status)
		if parseErr != nil {
			return nil, &ValidationError{Message: parseErr.Error()}
		}
		task.Status = newStatus
	}

	if req.DueDate != nil {
		newDue, parseErr := utils.ParseFutureDate(*req.DueDate)
		if parseErr != nil {
			return nil, &ValidationError{Message: parseErr.Error()}
		}
		task.DueDate = newDue
	}

	if err := s.repo.Update(task); err != nil {
		return nil, err
	}
	return task, nil
}

// DeleteTask removes the task identified by id.
func (s *taskService) DeleteTask(id string) error {
	return s.repo.Delete(id)
}

// ListTasks returns all tasks filtered by status (optional) and sorted by
// due_date ascending, with page/limit pagination applied.
func (s *taskService) ListTasks(opts ListOptions) ([]*domain.Task, int, error) {
	all, err := s.repo.GetAll()
	if err != nil {
		return nil, 0, err
	}

	// Filter by status when requested.
	if opts.Status != "" {
		if _, parseErr := domain.ParseStatus(opts.Status); parseErr != nil {
			return nil, 0, &ValidationError{Message: parseErr.Error()}
		}
		filtered := all[:0]
		for _, t := range all {
			if string(t.Status) == opts.Status {
				filtered = append(filtered, t)
			}
		}
		all = filtered
	}

	// Sort ascending by due_date.
	sort.Slice(all, func(i, j int) bool {
		return all[i].DueDate.Before(all[j].DueDate)
	})

	total := len(all)

	// Apply pagination.
	if opts.Limit > 0 {
		start := (opts.Page - 1) * opts.Limit
		if start < 0 {
			start = 0
		}
		if start >= total {
			return []*domain.Task{}, total, nil
		}
		end := start + opts.Limit
		if end > total {
			end = total
		}
		all = all[start:end]
	}

	return all, total, nil
}

// ValidationError signals that the caller supplied invalid input.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

// IsValidationError reports whether err is a ValidationError.
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// IsNotFound reports whether err represents a missing resource.
func IsNotFound(err error) bool {
	return err == apperrors.ErrTaskNotFound
}
