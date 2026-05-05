package dto

import "github.com/bhaskar2610/india-satcom-coding-challange/internal/domain"

// TaskResponse is the JSON representation returned to callers.
type TaskResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	DueDate     string `json:"due_date"`
}

// TaskListResponse wraps a list of tasks together with pagination metadata.
type TaskListResponse struct {
	Tasks      []TaskResponse `json:"tasks"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Error string `json:"error"`
}

const dueDateLayout = "2006-01-02"

// FromDomain converts a domain Task to a TaskResponse DTO.
func FromDomain(t *domain.Task) TaskResponse {
	return TaskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		DueDate:     t.DueDate.Format(dueDateLayout),
	}
}
