package handlers

import (
	"78/database"
	"78/logger"
	"78/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)

}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})

}
func buildTaskQuery(status, sort, order string) (string, []any) {
	query := `SELECT id, title, description, status, created_at, updated_at FROM task WHERE deleted_at IS NULL`
	args := []any{}
	argPos := 1

	// filter by status
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, status)
		argPos++
	}

	// sort column — whitelist allowed columns
	sortColumn := "created_at" // default
	switch sort {
	case "title":
		sortColumn = "title"
	case "status":
		sortColumn = "status"
	case "updated_at":
		sortColumn = "updated_at"
	case "created_at", "":
		sortColumn = "created_at"
	}

	// sort direction — whitelist asc/desc
	sortOrder := "DESC" // default
	switch strings.ToLower(order) {
	case "asc":
		sortOrder = "ASC"
	case "desc", "":
		sortOrder = "DESC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortColumn, sortOrder)

	return query, args
}

func CreateTask(w http.ResponseWriter, r *http.Request) {
	var dto models.CreateTaskDTO

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "invalid Json Body")
		return
	}

	// map DTO -> internal model
	task := models.Task{
		Title:       dto.Title,
		Description: dto.Description,
		Status:      dto.Status,
	}

	if errs := task.CreateValidate(); len(errs) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"errors": errs,
		})
		return
	}

	db := database.GetDB()

	row := db.QueryRow(
		`INSERT INTO task(title, description, status)
				VALUES ($1, $2, $3)
				RETURNING id, title, description, status, created_at, updated_at`,
		task.Title,
		task.Description,
		task.Status,
	)

	if err := row.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to Create task")
		return
	}

	writeJSON(w, http.StatusCreated, task.ToResponseDTO())
	logger.Log.Debug("task created")
}

func GetAllTasks(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()

	status := r.URL.Query().Get("status")
	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order") // asc / desc

	query, args := buildTaskQuery(status, sort, order)

	rows, err := db.Query(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get all tasks")
		return
	}
	defer rows.Close()

	tasksDTO := []models.TaskResponseDTO{}
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get all tasks")
			return
		}
		tasksDTO = append(tasksDTO, task.ToResponseDTO())
	}

	writeJSON(w, http.StatusOK, tasksDTO)
	logger.Log.Debug("got all tasks")
}

func GetTaskByID(w http.ResponseWriter, r *http.Request) {

	id := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Invalid Task ID")
		return
	}

	db := database.GetDB()
	var task models.Task

	err := db.QueryRow(
		`SELECT id, title, description, status, created_at, updated_at 
		 FROM task 
		 WHERE id = $1 AND deleted_at IS NULL`, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	writeJSON(w, http.StatusOK, task.ToResponseDTO())
	logger.Log.Debug("got task by id")
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {

	id := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Invalid Task ID")
		return
	}

	var dto models.UpdateTaskDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "invalid Json Body")
		return
	}

	task := models.Task{
		Title:       dto.Title,
		Description: dto.Description,
		Status:      dto.Status,
	}

	if errs := task.ValidateUpdate(); len(errs) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"errors": errs,
		})
		return
	}

	db := database.GetDB()

	row := db.QueryRow(
		`UPDATE task
				SET title = $1, description = $2, status = $3
				WHERE id = $4 AND deleted_at IS NULL
				RETURNING id, title, description, status, created_at, updated_at`,
		task.Title,
		task.Description,
		task.Status,
		id,
	)

	if err := row.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update task")
		return
	}

	writeJSON(w, http.StatusOK, task.ToResponseDTO())
	logger.Log.Debug("task updated")
}

func DeleteTaskHard(w http.ResponseWriter, r *http.Request) {

	id := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Invalid Task ID")
		return

	}
	db := database.GetDB()

	result, err := db.Exec(`DELETE FROM task WHERE id = $1`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete task")
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, http.StatusNotFound, "Task not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"result": "Task deleted"})
	logger.Log.Debug("task deleted")

}

func DeleteTaskSoft(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Invalid Task ID")
		return
	}

	db := database.GetDB()

	result, err := db.Exec(
		`UPDATE task SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete task")
		return
	}

	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, http.StatusNotFound, "Task not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"result": "Task soft-deleted"})
	logger.Log.Debug("task soft deleted")
}
