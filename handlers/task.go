package handlers

import (
	"78/database"
	"78/logger"
	"78/models"
	"encoding/json"
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

func CreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeError(w, http.StatusBadRequest, "invalid Json Body")
		return
	}

	if errs := task.CreateValidate(); len(errs) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"errors": errs,
		})
		return
	}

	db := database.GetDB()

	row := db.QueryRow(
		`INSERT INTO task(title ,description, status)
				VALUES ($1, $2, $3)
				Returning id , title, description, status, created_at, updated_at`,
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
	writeJSON(w, http.StatusCreated, task)
	logger.Log.Debug("task created")

}

func GetAllTasks(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	rows, err := db.Query(`SELECT id, title, description, status, created_at, updated_at FROM task ORDER BY created_at DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get all tasks")
		return
	}
	defer rows.Close()

	tasks := []models.Task{}
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
		tasks = append(tasks, task)
	}

	writeJSON(w, http.StatusOK, tasks)
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
		`SELECT id, title, description, status, created_at, updated_at FROM task WHERE id = $1`, id).Scan(
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

	writeJSON(w, http.StatusOK, task)
	logger.Log.Debug("got task by id")
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {

	id := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Invalid Task ID")
		return
	}

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeError(w, http.StatusBadRequest, "invalid Json Body")
		return

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
				WHERE id = $4
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

	writeJSON(w, http.StatusOK, task)
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
