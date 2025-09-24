package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stlesnik/url_shortener/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestLogging(t *testing.T) {
	// Initialize logger
	err := logger.InitLogger("dev")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("loggingResponseWriter", func(t *testing.T) {
		rr := httptest.NewRecorder()
		lw := &loggingResponseWriter{ResponseWriter: rr}

		lw.WriteHeader(http.StatusOK)
		assert.Equal(t, http.StatusOK, lw.status)

		size, err := lw.Write([]byte("hello"))
		assert.NoError(t, err)
		assert.Equal(t, 5, size)
		assert.Equal(t, 5, lw.size)
	})

	t.Run("WithLogging", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, err := w.Write([]byte("test"))
			assert.NoError(t, err)
		})

		WithLogging(next).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.Equal(t, "test", rr.Body.String())
	})
}
