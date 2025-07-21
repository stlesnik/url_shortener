package middleware

import (
	"net/http"
	"time"

	"github.com/stlesnik/url_shortener/internal/logger"
)

// loggingResponseWriter wraps http.ResponseWriter to capture status and size for logging.
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// Write writes the response body and updates the size for logging.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}

// WriteHeader writes the HTTP status code and updates the status for logging.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.status = statusCode
}

// WithLogging is a middleware that logs HTTP requests and responses.
func WithLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Sugaarz.Infow("Got request",
			"uri", r.RequestURI,
			"method", r.Method,
		)
		start := time.Now()

		lw := loggingResponseWriter{
			ResponseWriter: w,
		}
		next(&lw, r)

		duration := time.Since(start)

		logger.Sugaarz.Infow("Sent response",
			"status", lw.status,
			"size", lw.size,
			"duration", duration,
		)
	}
}
