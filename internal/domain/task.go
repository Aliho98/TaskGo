package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

type Task struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
type TaskStatus string

const (
	TaskStatusNew       TaskStatus = "pending"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusRunning   TaskStatus = "running"
)

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*Task, error)
	List(ctx context.Context, limit, offset int) ([]*Task, int64, error)
	Update(ctx context.Context, task *Task) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	HardDelete(ctx context.Context, id uuid.UUID) error
}

//func (t Task) ToResponseDTO() models.TaskResponseDTO {
//	return models.TaskResponseDTO{
//		ID:          t.ID,
//		Title:       t.Title,
//		Description: t.Description,
//		Status:      t.Status,
//		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
//		UpdatedAt:   t.UpdatedAt.Format(time.RFC3339),
//	}
//}
