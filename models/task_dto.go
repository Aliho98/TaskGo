package models

type CreateTaskDTO struct {
	Title       string `json:"title" validate:"required,min=3,max=255"`
	Description string `json:"description"`
	Status      string `json:"status" validate:"omitempty,oneof=pending in_progress done"`
}
type UpdateTaskDTO struct {
	Title       string `json:"title" validate:"required,min=3,max=255"`
	Description string `json:"description"`
	Status      string `json:"status" validate:"omitempty,oneof=pending in_progress done"`
}
type TaskResponseDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
