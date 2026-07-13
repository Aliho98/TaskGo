package postgres

import (
	"78/internal/domain"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
)

type TaskRow struct {
	ID          uuid.UUID    `db:"id"`
	Title       string       `db:"title"`
	Description string       `db:"description"`
	Status      string       `db:"status"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
	DeletedAt   sql.NullTime `db:"deleted_at"`
}

func (r *TaskRow) TableName() *domain.Task {
	t := &domain.Task{
		ID:          r.ID,
		Title:       r.Title,
		Description: r.Description,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
	if r.DeletedAt.Valid {
		t.DeletedAt = &r.DeletedAt.Time
	}
	return t
}

const taskTable = "tasks"

type TaskRepository struct {
	db *goqu.Database
}

func NewTaskRepository(sqlDB *sql.DB) *TaskRepository {
	return &TaskRepository{db: goqu.New("postgres", sqlDB)}
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	_, err := r.db.Insert(taskTable).Rows(goqu.Record{
		"id":          task.ID,
		"title":       task.Title,
		"description": task.Description,
		"status":      task.Status,
		"created_at":  task.CreatedAt,
		"updated_at":  task.UpdatedAt,
	}).Executor().ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("Error inserting task: %v", err)
	}
	return nil
}

func (r *TaskRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	var row TaskRow

	found, err := r.db.From(taskTable).
		Where(goqu.Ex{"id": id, "deleted_at": nil}).
		ScanStructContext(ctx, &row)
	if err != nil {
		return nil, fmt.Errorf("Error getting task: %v", err)
	}
	if !found {
		return nil, domain.ErrTaskNotFound
	}
	return row.TableName(), nil
}

func (r *TaskRepository) List(ctx context.Context, limit, offset int) ([]*domain.Task, int64, error) {
	var rows []TaskRow

	err := r.db.From(taskTable).
		Where(goqu.Ex{"deleted_at": nil}).
		Order(goqu.I("created_at ").Desc()).
		Limit(uint(limit)).
		Offset(uint(offset)).
		ScanStructsContext(ctx, &rows)
	if err != nil {
		return nil, 0, fmt.Errorf("List task: %v", err)
	}

	var total int64

	_, err = r.db.From(taskTable).
		Select(goqu.Count("*")).
		Where(goqu.Ex{"deleted_at": nil}).
		ScanValContext(ctx, &total)
	if err != nil {
		return nil, 0, fmt.Errorf("Count task: %v", err)

	}

	tasks := make([]*domain.Task, len(rows))
	for _, row := range rows {
		tasks = append(tasks, row.toDomain())
	}
}
