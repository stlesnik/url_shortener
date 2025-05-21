package handlers

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stlesnik/url_shortener/internal/app/middleware"
	"github.com/stlesnik/url_shortener/internal/app/models"
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

	service := services.New(repo, cfg)
	handler := New(service)

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
	service := services.New(repo, cfg)
	handler := New(service)

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
			ctx := context.WithValue(r.Context(), middleware.UserIDKeyName, "test")
			r = r.WithContext(ctx)
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
			SaveURL(gomock.Any(), gomock.Any(), longURL, "").
			Return(false, nil).
			Times(1),
		m.EXPECT().
			SaveURL(gomock.Any(), gomock.Any(), longURL, "").
			Return(true, nil).
			Times(1),
	)

	cfg := &config.Config{BaseURL: "http://localhost:8000"}
	err := logger.InitLogger(cfg.Environment)
	require.NoError(t, err)
	service := services.New(m, cfg)
	handler := New(service)

	req1 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(longURL))
	req1.Header.Add("Content-Type", "text/plain")
	ctx1 := context.WithValue(req1.Context(), middleware.UserIDKeyName, "")
	w1 := httptest.NewRecorder()
	handler.SaveURL(w1, req1.WithContext(ctx1))

	require.Equal(t, http.StatusCreated, w1.Code)
	assert.Equal(t, "text/plain", w1.Header().Get("Content-Type"))
	res1 := w1.Result()
	body1, err := io.ReadAll(res1.Body)
	require.NoError(t, err)
	err = res1.Body.Close()
	require.NoError(t, err)

	req2 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(longURL))
	req2.Header.Add("Content-Type", "text/plain")
	ctx2 := context.WithValue(req2.Context(), middleware.UserIDKeyName, "")
	w2 := httptest.NewRecorder()
	handler.SaveURL(w2, req2.WithContext(ctx2))

	require.Equal(t, http.StatusConflict, w2.Code)
	assert.Equal(t, "text/plain", w2.Header().Get("Content-Type"))
	res2 := w2.Result()
	body2, err := io.ReadAll(res2.Body)
	require.NoError(t, err)
	err = res2.Body.Close()
	require.NoError(t, err)
	assert.Equal(t, body1, body2)

	err = w1.Result().Body.Close()
	if err != nil {
		panic(err)
	}
	err = w2.Result().Body.Close()
	if err != nil {
		panic(err)
	}
}

func TestHandler_GetLongURL(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8000"} // Добавляем конфиг
	err := logger.InitLogger(cfg.Environment)
	require.NoError(t, err)
	repo := repository.NewInMemoryRepository()
	_, _ = repo.SaveURL(context.Background(), "_SGMGLQIsIM=", "http://mbrgaoyhv.yandex", "")
	service := services.New(repo, cfg)
	handler := New(service)

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
	service := services.New(repo, cfg)
	handler := New(service)

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

func TestHandler_APIGetUserURLs(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8000"}
	_ = logger.InitLogger(cfg.Environment)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type (
		FullRepo struct {
			*mocks.MockRepository
			*mocks.MockURLList
		}
	)

	tests := []struct {
		name         string
		setupRepo    func() services.Repository
		setupContext func(*http.Request) *http.Request
		expectCall   func(*FullRepo)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Репо поддерживает URLList - успех",
			setupRepo: func() services.Repository {
				fr := &FullRepo{
					mocks.NewMockRepository(ctrl),
					mocks.NewMockURLList(ctrl),
				}
				return fr
			},
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), middleware.UserIDKeyName, "user123"))
			},
			expectCall: func(fr *FullRepo) {
				fr.MockURLList.EXPECT().
					GetURLList(gomock.Any(), "user123").
					Return([]models.BaseURLResponse{
						{ShortURL: "http://localhost/abc", OriginalURL: "https://ya.ru"},
					}, nil)
			},
			expectedCode: http.StatusCreated,
			expectedBody: `[{"short_url":"http://localhost/abc","original_url":"https://ya.ru"}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			if fr, ok := repo.(*FullRepo); ok && tt.expectCall != nil {
				tt.expectCall(fr)
			}

			service := services.New(repo, cfg)
			handler := New(service)

			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			if tt.setupContext != nil {
				req = tt.setupContext(req)
			}
			w := httptest.NewRecorder()

			handler.APIGetUserURLs(w, req)

			res := w.Result()
			defer func() {
				require.NoError(t, req.Body.Close())
			}()

			assert.Equal(t, tt.expectedCode, res.StatusCode)
			if tt.expectedBody != "" {
				body, _ := io.ReadAll(res.Body)
				assert.JSONEq(t, tt.expectedBody, string(body))
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
	m.EXPECT().Ping(context.Background()).Return(nil)
	require.NoError(t, err)
	service := services.New(m, cfg)
	handler := New(service)

	t.Run("Mock test db", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()
		handler.PingDB(w, r)

		require.Equal(t, 200, w.Code)
	})

}
