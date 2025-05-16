package handlers

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stlesnik/url_shortener/internal/app/repository"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"github.com/stlesnik/url_shortener/internal/app/services/mocks"
	"github.com/stlesnik/url_shortener/internal/config"
	"github.com/stlesnik/url_shortener/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_getLongURLFromReq(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8000"}
	err := logger.InitLogger(cfg.Environment)
	require.NoError(t, err)
	repo := repository.NewInMemoryRepository()

	service := services.NewURLShortenerService(repo, cfg)
	handler := NewHandler(service)

	type expected struct {
		longURLStr string
		error      string
	}
	tests := []struct {
		name     string
		longURL  string
		expected expected
	}{
		{
			name:     "good case",
			longURL:  "http://mbrgaoyhv.yandex",
			expected: expected{longURLStr: "http://mbrgaoyhv.yandex", error: ""},
		},
		{
			name:     "bad body",
			longURL:  ``,
			expected: expected{longURLStr: "", error: "error getting url"},
		},
		{
			name:     "bad url",
			longURL:  "://mbrgaoyhv.yandex",
			expected: expected{longURLStr: "", error: "got incorrect url to shorten: url=://mbrgaoyhv.yandex, err=got incorrect url to shorten: url=://mbrgaoyhv.yandex, err= parse \"://mbrgaoyhv.yandex\": missing protocol scheme: invalid url to shorten"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.longURL))
			req.Header.Add("Content-Type", "text/plain")

			longURLStr, err := handler.getLongURLFromReq(req)

			if tt.expected.error != "" {
				require.EqualError(t, err, tt.expected.error)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.longURLStr, longURLStr)
			}
		})
	}
}

func TestHandler_SaveURL(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8000"}
	err := logger.InitLogger(cfg.Environment)
	require.NoError(t, err)
	repo := repository.NewInMemoryRepository()
	service := services.NewURLShortenerService(repo, cfg)
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
				err := res.Body.Close()
				if err != nil {
					panic(err)
				}
			}
		})
	}
}

func TestHandler_SaveURL_Conflict_WithMockRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockRepository(ctrl)

	const longURL = "http://example.com"
	gomock.InOrder(
		m.EXPECT().
			Save(gomock.Any(), longURL).
			Return(false, nil).
			Times(1),
		m.EXPECT().
			Save(gomock.Any(), longURL).
			Return(true, nil).
			Times(1),
	)

	cfg := &config.Config{BaseURL: "http://localhost:8000"}
	err := logger.InitLogger(cfg.Environment)
	require.NoError(t, err)
	service := services.NewURLShortenerService(m, cfg)
	handler := NewHandler(service)

	req1 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(longURL))
	req1.Header.Add("Content-Type", "text/plain")
	w1 := httptest.NewRecorder()
	handler.SaveURL(w1, req1)

	require.Equal(t, http.StatusCreated, w1.Code)
	assert.Equal(t, "text/plain", w1.Header().Get("Content-Type"))
	body1, _ := io.ReadAll(w1.Result().Body)

	req2 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(longURL))
	req2.Header.Add("Content-Type", "text/plain")
	w2 := httptest.NewRecorder()
	handler.SaveURL(w2, req2)

	require.Equal(t, http.StatusConflict, w2.Code)
	assert.Equal(t, "text/plain", w2.Header().Get("Content-Type"))
	body2, _ := io.ReadAll(w2.Result().Body)
	assert.Equal(t, body1, body2)
}

func TestHandler_GetLongURL(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8000"} // Добавляем конфиг
	err := logger.InitLogger(cfg.Environment)
	require.NoError(t, err)
	repo := repository.NewInMemoryRepository()
	_, _ = repo.Save("_SGMGLQIsIM=", "http://mbrgaoyhv.yandex")
	service := services.NewURLShortenerService(repo, cfg)
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

func TestHandler_ApiPrepareShortURL(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8000"}
	err := logger.InitLogger(cfg.Environment)
	require.NoError(t, err)
	repo := repository.NewInMemoryRepository()
	service := services.NewURLShortenerService(repo, cfg)
	handler := NewHandler(service)

	tests := []struct {
		name         string
		body         string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Good json case",
			body:         `{"url":"https://vk.com"}`,
			expectedCode: 201,
			expectedBody: `{"result":"http://localhost:8000/ymMooIzfwh4="}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.body))
			r.Header.Add("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.APIPrepareShortURL(w, r)

			require.Equal(t, tt.expectedCode, w.Code)
			res := w.Result()
			if tt.expectedCode == http.StatusCreated {
				resJSONBytes, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				resJSON := string(resJSONBytes)
				assert.JSONEq(t, tt.expectedBody, resJSON)
			}
			err := res.Body.Close()
			require.NoError(t, err)
		})
	}
}

func TestHandler_PingDB(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8000"} // Добавляем конфиг
	err := logger.InitLogger(cfg.Environment)
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockRepository(ctrl)
	m.EXPECT().Ping().Return(nil)
	require.NoError(t, err)
	service := services.NewURLShortenerService(m, cfg)
	handler := NewHandler(service)

	t.Run("Mock test db", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()
		handler.PingDB(w, r)

		require.Equal(t, 200, w.Code)
	})

}
