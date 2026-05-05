package dto

// CreateTaskRequest is the payload accepted by POST /tasks.
type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	DueDate     string `json:"due_date"`
}

// UpdateTaskRequest is the payload accepted by PUT /tasks/{id}.
// All fields are pointers so that a missing field is distinguishable from an
// explicit zero value (partial-update semantics).
type UpdateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	DueDate     *string `json:"due_date"`
}
