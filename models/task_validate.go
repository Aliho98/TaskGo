package models

import (
	"78/internal/domain"
	"strings"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (t *domain.Task) CreateValidate() []ValidationError {
	var errs []ValidationError

	if strings.TrimSpace(t.Title) == "" {
		errs = append(errs, ValidationError{"title", "title is required"})
	} else if len(t.Title) > 255 {
		errs = append(errs, ValidationError{"title", "title must be under 255 characters"})
	}

	if len(t.Description) > 1000 {
		errs = append(errs, ValidationError{"description", "description must be under 1000 characters"})
	}

	validStatuses := map[string]bool{"pending": true, "in_progress": true, "done": true}
	if t.Status == "" {
		t.Status = "pending" // set default
	} else if !validStatuses[t.Status] {
		errs = append(errs, ValidationError{"status", "status must be pending, in_progress, or done"})
	}

	return errs
}

func (t *domain.Task) ValidateUpdate() []ValidationError {
	var errs []ValidationError

	if t.Title != "" && len(t.Title) > 255 {
		errs = append(errs, ValidationError{"title", "title must be under 255 characters"})
	}

	if len(t.Description) > 1000 {
		errs = append(errs, ValidationError{"description", "description must be under 1000 characters"})
	}

	validStatuses := map[string]bool{"pending": true, "in_progress": true, "done": true}
	if t.Status != "" && !validStatuses[t.Status] {
		errs = append(errs, ValidationError{"status", "status must be pending, in_progress, or done"})
	}

	return errs
}
