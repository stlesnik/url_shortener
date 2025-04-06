package handlers

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/stlesnik/url_shortener/cmd/config"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"github.com/stlesnik/url_shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_SaveURL(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8000"}
	repo := storage.NewInMemoryStorage()
	service := services.NewUrlShortenerService(repo, cfg)
	handler := NewHandler(service)

	type expected struct {
		contentType string
		statusCode  int
		body        string
	}

	tests := []struct {
		name     string
		longURL  string
		expected expected
	}{
		{
			name:    "valid url",
			longURL: "http://mbrgaoyhv.yandex",
			expected: expected{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
				body:        "http://localhost:8000/_SGMGLQIsIM=",
			},
		},
		{
			name:    "invalid url",
			longURL: "invalid-url",
			expected: expected{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
				body:        "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.longURL))
			r.Header.Add("Content-Type", "text/plain")
			w := httptest.NewRecorder()
			handler.SaveURL(w, r)

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

func TestHandler_GetLongURL(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8000"} // Добавляем конфиг
	repo := storage.NewInMemoryStorage()
	repo.Save("_SGMGLQIsIM=", "http://mbrgaoyhv.yandex")
	service := services.NewUrlShortenerService(repo, cfg)
	handler := NewHandler(service)

	type expected struct {
		statusCode int
		location   string
	}

	tests := []struct {
		name     string
		path     string
		expected expected
	}{
		{
			name: "Url exists",
			path: "/_SGMGLQIsIM=",
			expected: expected{
				statusCode: http.StatusTemporaryRedirect,
				location:   "http://mbrgaoyhv.yandex",
			},
		},
		{
			name: "Url dont exist",
			path: "/invalid-path",
			expected: expected{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.path, nil)
			//prepare chi context
			id := strings.TrimPrefix(tt.path, "/")
			rc := chi.NewRouteContext()
			rc.URLParams.Add("id", id)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rc))

			w := httptest.NewRecorder()
			handler.GetLongURL(w, r)

			require.Equal(t, tt.expected.statusCode, w.Code)
			if tt.expected.statusCode == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.expected.location, w.Header().Get("Location"))
			}
		})
	}
}
