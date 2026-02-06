package middleware

import (
	"NeoBIT/internal/metrics"
	"net/http"
	"time"
)

func Metrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			metrics.IncActive(r.URL.Path)
			defer metrics.DecActive(r.URL.Path)

			start := time.Now()
			rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			duration := time.Since(start)
			metrics.ObserveRequest(r.Method, r.URL.Path, rec.status, duration)
		})
	}
}
