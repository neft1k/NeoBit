package middleware

import (
	"net"
	"net/http"
	"strings"

	"NeoBIT/internal/logger"
	"NeoBIT/internal/metrics"
	"NeoBIT/internal/ratelimit"
)

func RateLimiter(capacity, refillRate int, log logger.Logger) func(http.Handler) http.Handler {
	if log == nil {
		log = logger.Nop()
	}
	limiter := ratelimit.NewRateLimiter(capacity, refillRate)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := clientKey(r)
			if !limiter.Allow(key) {
				metrics.IncRateLimitExceeded()
				log.Warn(r.Context(), "rate limit exceeded",
					logger.FieldAny("client", key),
					logger.FieldAny("method", r.Method),
					logger.FieldAny("path", r.URL.Path),
				)
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientKey(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
