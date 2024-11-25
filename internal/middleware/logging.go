// internal/middleware/logging.go

package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		// Log after request is complete
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_ip", r.RemoteAddr).
			Dur("duration", time.Since(start)).
			Msg("Request processed")
	})
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a wrapper around ResponseWriter to capture the status code
		wrapper := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(wrapper, r)

		// Calculate duration
		duration := time.Since(start)

		// Log the request
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_ip", r.RemoteAddr).
			Int("status", wrapper.statusCode).
			Dur("duration", duration).
			Str("user_agent", r.UserAgent()).
			Msg("Request processed")
	})
}

// responseWriter is a wrapper for http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
