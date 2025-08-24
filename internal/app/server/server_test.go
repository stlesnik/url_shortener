package server

import (
	"github.com/stlesnik/url_shortener/internal/logger"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stlesnik/url_shortener/internal/app/services/mocks"
	"github.com/stlesnik/url_shortener/internal/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorager := mocks.NewMockStorager(ctrl)
	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}
	err := logger.InitLogger(cfg.Environment)
	require.NoError(t, err)
	daemonsDoneCh := make(chan struct{})

	t.Run("New", func(t *testing.T) {
		s, err := New(mockStorager, cfg, daemonsDoneCh)
		assert.NoError(t, err)
		assert.NotNil(t, s)
	})

	t.Run("New with nil repo", func(t *testing.T) {
		_, err := New(nil, cfg, daemonsDoneCh)
		assert.Error(t, err)
	})

	t.Run("New with nil cfg", func(t *testing.T) {
		_, err := New(mockStorager, nil, daemonsDoneCh)
		assert.Error(t, err)
	})

	t.Run("New with nil daemonsDoneCh", func(t *testing.T) {
		_, err := New(mockStorager, cfg, nil)
		assert.Error(t, err)
	})

	t.Run("Ping route", func(t *testing.T) {
		s, _ := New(mockStorager, cfg, daemonsDoneCh)
		mockStorager.EXPECT().Ping(gomock.Any()).Return(nil)

		req := httptest.NewRequest("GET", "/ping", nil)
		rr := httptest.NewRecorder()
		s.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
