package service

import (
	"78/pkg/pagination"
	"context"
	"fmt"

	"78/internal/domain"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type CreateTaskInput struct {
	Title       string
	Description string
}

type ListFilter struct {
	Status  *domain.TaskStatus
	SortBy  string
	SortDir string
}

type UpdateTaskInput struct {
	Title       *string
	Description *string
	Status      *domain.TaskStatus
}

type TaskService struct {
	repo domain.TaskRepository
	log  *zap.Logger
}

func NewTaskService(repo domain.TaskRepository, log *zap.Logger) *TaskService {
	return &TaskService{repo: repo, log: log}
}

func (s *TaskService) CreateTask(ctx context.Context, in CreateTaskInput) (*domain.Task, error) {
	task := &domain.Task{
		ID:          uuid.New(),
		Title:       in.Title,
		Description: in.Description,
		Status:      domain.TaskStatusRunning,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		s.log.Error("Failed to create task", zap.Error(err))
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	return task, nil
}

func (s *TaskService) GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) ListTasks(ctx context.Context, p pagination.Pagination, filter ListFilter) ([]*domain.Task, int64, error) {
	tasks, total, err := s.repo.List(ctx, domain.ListParams{
		Limit:   p.Limit(),
		Offset:  p.Offset(),
		Status:  filter.Status,
		SortBy:  filter.SortBy,
		SortDir: filter.SortDir,
	})
	if err != nil {
		s.log.Error("Failed to list tasks", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}
	return tasks, total, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, id uuid.UUID, update UpdateTaskInput) (*domain.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if update.Title != nil {
		task.Title = *update.Title
	}
	if update.Description != nil {
		task.Description = *update.Description
	}
	if update.Status != nil {
		task.Status = *update.Status
	}
	if err := s.repo.Update(ctx, task); err != nil {
		s.log.Error("Failed to update task", zap.Error(err))
		return nil, fmt.Errorf("failed to update task: %w", err)
	}
	return task, nil
}

func (s *TaskService) SoftDeleteTask(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("failed to soft delete task: %w", err)
	}
	return nil
}

func (s *TaskService) HardDeleteTask(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.HardDelete(ctx, id); err != nil {
		return fmt.Errorf("failed to hard delete task: %w", err)
	}
	return nil
}
