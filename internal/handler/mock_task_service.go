package handler

import (
	"github.com/stretchr/testify/mock"

	"github.com/bhaskar2610/india-satcom-coding-challange/internal/domain"
	"github.com/bhaskar2610/india-satcom-coding-challange/internal/dto"
	"github.com/bhaskar2610/india-satcom-coding-challange/internal/service"
)

// MockTaskService is a testify mock that satisfies service.TaskService.
type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateTask(req dto.CreateTaskRequest) (*domain.Task, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskService) GetTaskByID(id string) (*domain.Task, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskService) UpdateTask(id string, req dto.UpdateTaskRequest) (*domain.Task, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskService) DeleteTask(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTaskService) ListTasks(opts service.ListOptions) ([]*domain.Task, int, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.Task), args.Int(1), args.Error(2)
}
