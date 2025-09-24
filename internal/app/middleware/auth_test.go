package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stlesnik/url_shortener/internal/config"
	"github.com/stlesnik/url_shortener/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	err := logger.InitLogger("dev")
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		AuthSecretKey: "test_secret_key",
	}

	t.Run("createSignedCookie", func(t *testing.T) {
		cookie, err := createSignedCookie("user1", cfg.AuthSecretKey)
		assert.NoError(t, err)
		assert.NotNil(t, cookie)
		assert.Equal(t, "auth_token", cookie.Name)
	})

	t.Run("getUserIDFromCookie", func(t *testing.T) {
		cookie, _ := createSignedCookie("user1", cfg.AuthSecretKey)
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(cookie)

		userID, err := getUserIDFromCookie(req, cfg.AuthSecretKey)
		assert.NoError(t, err)
		assert.Equal(t, "user1", userID)
	})

	t.Run("getUserIDFromCookie no cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		_, err := getUserIDFromCookie(req, cfg.AuthSecretKey)
		assert.Error(t, err)
	})

	t.Run("WithAuth no cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(UserIDKeyName)
			assert.NotNil(t, userID)
		})

		WithAuth(cfg, next).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NotEmpty(t, rr.Header().Get("Set-Cookie"))
	})

	t.Run("WithAuth with cookie", func(t *testing.T) {
		cookie, _ := createSignedCookie("user1", cfg.AuthSecretKey)
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(cookie)
		rr := httptest.NewRecorder()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(UserIDKeyName)
			assert.NotNil(t, userID)
			assert.Equal(t, "user1", userID)
		})

		WithAuth(cfg, next).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Header().Get("Set-Cookie"))
	})

	t.Run("getUserIDFromCookie invalid cookie", func(t *testing.T) {
		cookie := &http.Cookie{
			Name:  "auth_token",
			Value: "invalid",
		}
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(cookie)
		_, err := getUserIDFromCookie(req, cfg.AuthSecretKey)
		assert.Error(t, err)
	})

	t.Run("createSignedCookie expiration", func(t *testing.T) {
		cookie, err := createSignedCookie("user1", cfg.AuthSecretKey)
		assert.NoError(t, err)
		assert.True(t, cookie.Expires.After(time.Now()))
	})
}
