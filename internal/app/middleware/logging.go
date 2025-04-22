package middleware

import (
	"github.com/stlesnik/url_shortener/cmd/logger"
	"net/http"
	"time"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func WithLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		next(&lw, r)

		duration := time.Since(start)
		logger.Sugaarz.Infow("Got request",
			"uri", r.RequestURI,
			"method", r.Method,
			"duration", duration,
		)
		logger.Sugaarz.Infow("Sent response",
			"status", responseData.status,
			"size", responseData.size,
		)
	}
}
