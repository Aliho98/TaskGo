package tests

import (
	"78/database"
	"78/handlers"
	"78/logger"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMain(m *testing.M) {
	logger.Init()
	m.Run()
}

func TestCreateTask_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer mockDB.Close()

	// swap real DB with mock
	database.DB = mockDB

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "title", "description", "status", "created_at", "updated_at"}).
		AddRow("uuid-123", "Test Task", "desc", "pending", now, now)

	mock.ExpectQuery("INSERT INTO task").
		WithArgs("Test Task", "desc", "pending").
		WillReturnRows(rows)

	body := bytes.NewBufferString(`{"title":"Test Task","description":"desc","status":"pending"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	w := httptest.NewRecorder()

	handlers.CreateTask(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", res.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateTask_InvalidJSON(t *testing.T) {
	body := bytes.NewBufferString(`{invalid`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	w := httptest.NewRecorder()

	handlers.CreateTask(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", res.StatusCode)
	}
}

func TestCreateTask_ValidationFails(t *testing.T) {
	// assumes CreateValidate() requires Title to be non-empty
	body := bytes.NewBufferString(`{"title":"","description":"desc","status":"pending"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	w := httptest.NewRecorder()

	handlers.CreateTask(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", res.StatusCode)
	}
}
