package handlers

import (
	"github.com/stlesnik/url_shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_processPostRequest(t *testing.T) {
	repo := storage.NewInMemoryStorage()
	handler := NewHandler(repo)

	type expected struct {
		contentType string
		statusCode  int
		body        string
	}

	tests := []struct {
		name        string
		requestBody string
		expected    expected
	}{
		{
			name:        "valid url",
			requestBody: "http://mbrgaoyhv.yandex",
			expected: expected{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
				body:        "http://example.com/_SGMGLQIsIM=",
			},
		},
		{
			name:        "invalid url",
			requestBody: "invalid-url",
			expected: expected{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
				body:        "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()
			handler.processPostRequest(w, r)

			require.Equal(t, tt.expected.statusCode, w.Code)
			res := w.Result()
			if tt.expected.statusCode == http.StatusCreated {
				assert.Equal(t, tt.expected.contentType, res.Header.Get("Content-Type"))

				shortURL, _ := io.ReadAll(res.Body)
				assert.Equal(t, tt.expected.body, string(shortURL))
				res.Body.Close()
			}
		})
	}
}

func TestHandler_processGetRequest(t *testing.T) {
	repo := storage.NewInMemoryStorage()
	repo.Save("_SGMGLQIsIM=", "http://mbrgaoyhv.yandex")
	handler := NewHandler(repo)

	type expected struct {
		statusCode int
		location   string
	}

	tests := []struct {
		name     string
		id       string
		expected expected
	}{
		{
			name: "Url exists",
			id:   "_SGMGLQIsIM=",
			expected: expected{
				statusCode: http.StatusTemporaryRedirect,
				location:   "http://mbrgaoyhv.yandex",
			},
		},
		{
			name: "Url dont exist",
			id:   "invalid-id",
			expected: expected{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			handler.processGetRequest(w, tt.id)

			require.Equal(t, tt.expected.statusCode, w.Code)
			res := w.Result()
			if tt.expected.statusCode == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.expected.location, res.Header.Get("Location"))
			}
		})
	}
}
