package dto

import (
	"78/internal/domain"
	"78/internal/service"
	"time"

	"github.com/google/uuid"
)

type CreateTaskRequest struct {
	Title       string `json:title" validate:"required,min=3,max=255"`
	Description string `json:description" validate:"required,min=3,max=2550"`
}

func (r CreateTaskRequest) ToServiceInput() service.CreateTaskInput {
	return service.CreateTaskInput{Title: r.Title, Description: r.Description}
}

type UpdateTaskRequest struct {
	Title       *string `json:title" validate:"required,min=3,max=255"`
	Description *string `json:description" validate:"required,min=3,max=2550"`
	Status      *string `json:status" validate:"required,oneof=pending in_progress completed"`
}

func (r UpdateTaskRequest) ToServiceInput() service.UpdateTaskInput {
	input := service.UpdateTaskInput{Title: r.Title, Description: r.Description}
	if r.Status != nil {
		s := domain.TaskStatus(*r.Status)
		input.Status = &s
	}
	return input
}

type TaskResponse struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func FromDomain(task *domain.Task) TaskResponse {
	return TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func FromDomainList(tasks []*domain.Task) []TaskResponse {
	responses := make([]TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		responses = append(responses, FromDomain(task))
	}
	return responses
}
