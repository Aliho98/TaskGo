package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func Logging(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(ww, r)

			log.Info("http request",
				zap.String("method", r.Method),
				zap.String("url", r.URL.String()),
				zap.Int("status", ww.status),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}
