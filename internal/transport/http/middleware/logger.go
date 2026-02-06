package middleware

import (
	"NeoBIT/internal/logger"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func HTTPLogger(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			duration := time.Since(start)
			log.Info(r.Context(), "http_request",
				logger.FieldAny("method", r.Method),
				logger.FieldAny("path", r.URL.Path),
				logger.FieldAny("status", rec.status),
				logger.FieldAny("duration_ms", duration.Milliseconds()),
			)
		})
	}
}
