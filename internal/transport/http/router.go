package http

import (
	"78/internal/transport/http/handler"
	"78/internal/transport/http/middleware"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"go.uber.org/zap"
)

func NewRouter(taskHandler *handler.TaskHandler, logger *zap.Logger) http.Handler {
	router := chi.NewRouter()

	router.Use(chimw.RequestID)
	router.Use(chimw.Recoverer)
	router.Use(middleware.Logging(logger))
	router.Use(chimw.Timeout(30 * time.Second))

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	router.Route("/task", func(r chi.Router) {
		r.Post("/", taskHandler.Create)
		r.Get("/{id}", taskHandler.Get)
		r.Get("/", taskHandler.List)
		r.Get("/{id}", taskHandler.Delete)
		r.Patch("/{id}", taskHandler.Update)
	})
	return router
}
