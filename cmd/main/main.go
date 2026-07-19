package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"78/internal/config"
	"78/internal/platform/database"
	"78/internal/platform/logger"
	"78/internal/repository/postgres"
	"78/internal/service"
	transporthttp "78/internal/transport/http" // aliased: avoids clashing with the imported "net/http" package
	"78/internal/transport/http/handler"
	"78/pkg/validator"

	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()

	log, err := logger.New(cfg.Env)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.NewPostgresDb(ctx, cfg.PostgresHost)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	taskRepo := postgres.NewTaskRepository(db)
	taskservice := service.NewTaskService(taskRepo, log)
	v := validator.New()
	taskHandler := handler.NewTaskHandler(taskservice, log, v)
	router := transporthttp.NewRouter(taskHandler, log)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Info("Starting server on port " + cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to listen on port " + cfg.HTTPPort)
		}
	}()
	<-ctx.Done()
	log.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Failed to shutdown server", zap.Error(err))
	}
	log.Info("Server gracefully stopped")

}
