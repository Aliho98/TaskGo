package handler

import (
	"78/internal/service"
	"78/internal/transport/http/dto"
	"78/pkg/validator"
	"encoding/json"
	"net/http"

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
