package service

import (
	"github.com/stretchr/testify/mock"

	"github.com/bhaskar2610/india-satcom-coding-challange/internal/domain"
)

// MockTaskRepository is a testify mock that satisfies repository.TaskRepository.
// It lives in the service package so tests can use it without a circular import.
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Save(task *domain.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepository) GetByID(id string) (*domain.Task, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(task *domain.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTaskRepository) GetAll() ([]*domain.Task, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}
