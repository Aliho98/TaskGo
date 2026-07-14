package http

import (
	"78/internal/transport/http/handler"
	"78/internal/transport/http/middleware"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"time"
)
func NewRouter (taskHandler *handler.TaskHandler), log *zap.Logger) http.Handler {
	r := chi.NewRouter()
	
}


