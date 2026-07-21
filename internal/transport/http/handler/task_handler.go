package handler

import (
	"78/internal/domain"
	"78/internal/service"
	"78/internal/transport/http/dto"
	"78/pkg/pagination"
	"78/pkg/validator"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TaskHandler struct {
	service   *service.TaskService
	log       *zap.Logger
	validator *validator.Validator
}

func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, map[string]string{"error": message})
}

func respondValidationErrors(w http.ResponseWriter, err error) {
	respondJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
		"error":  "validation failed",
		"fields": validator.FormatValidationError(err),
	})
}

func NewTaskHandler(service *service.TaskService, logger *zap.Logger, v *validator.Validator) *TaskHandler {
	return &TaskHandler{
		service:   service,
		log:       logger,
		validator: v,
	}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTaskRequest

	if err := json.NewEncoder(w).Encode(req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		respondValidationErrors(w, err)
	}

	task, err := h.service.CreateTask(r.Context(), req.ToServiceInput())
	if err != nil {
		h.log.Error("Create task error", zap.Error(err))
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, dto.FromDomain(task))

}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	task, err := h.service.GetTask(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		h.log.Error("Get task error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, dto.FromDomain(task))

}

var allowedSortBy = map[string]bool{
	"created_at": true,
	"updated_at": true,
	"title":      true,
	"status":     true,
}
var allowedSortDir = map[string]bool{
	"asc":  true,
	"desc": true,
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	p := pagination.FromQuery(q)
	filter := service.ListFilter{
		SortBy:  "created_at",
		SortDir: "desc",
	}

	if raw := q.Get("status"); raw != "" {
		status := domain.TaskStatus(raw)
		switch status {
		case domain.TaskStatusCompleted, domain.TaskStatusRunning, domain.TaskStatusNew:
			filter.Status = &status
		default:
			respondError(w, http.StatusBadRequest, "invalid status")
			return
		}

		if raw := q.Get("sort_by"); raw != "" {
			if !allowedSortBy[raw] {
				respondError(w, http.StatusBadRequest, "invalid sort by")
				return
			}
			filter.SortBy = raw
		}
		task, total, err := h.service.ListTasks(r.Context(), p, filter)
		if err != nil {
			h.log.Error("List task error", zap.Error(err))
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"data": dto.FromDomainList(task),
			"meta": pagination.NewMeta(p, total)})
	}
	
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	hard := r.URL.Query().Get("hard") == "true"
	if hard {
		err = h.service.HardDeleteTask(r.Context(), id)
	} else {
		err = h.service.SoftDeleteTask(r.Context(), id)
	}

	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
		}
		h.log.Error("Delete task error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)

}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	var req dto.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		respondValidationErrors(w, err)
		return
	}
	task, err := h.service.UpdateTask(r.Context(), id, req.ToServiceInput())
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		h.log.Error("Update task error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, err.Error())
		return

	}
	respondJSON(w, http.StatusOK, dto.FromDomain(task))
}
