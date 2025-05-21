package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/stlesnik/url_shortener/internal/config"
	"github.com/stlesnik/url_shortener/internal/logger"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type contextKey string

const (
	TOKEN_EXP                   = time.Hour * 24
	USER_ID_KEY_NAME contextKey = "userID"
)

func WithAuth(cfg *config.Config, createIfNot bool, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserIDFromCookie(r, cfg.AuthSecretKey)

		if err == nil {
			logger.Sugaarz.Infow("Got user id from cookie", "userID", userID)
			ctx := context.WithValue(r.Context(), USER_ID_KEY_NAME, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else if createIfNot {
			if errors.Is(err, http.ErrNoCookie) {
				logger.Sugaarz.Errorw("error while validating token:", "err", err)
			}
			newUserID := uuid.New().String()
			logger.Sugaarz.Infow("No user id in cookie. Created new", "userID", newUserID)
			cookie, err := createSignedCookie(newUserID, cfg.AuthSecretKey)
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, cookie)
			ctx := context.WithValue(r.Context(), USER_ID_KEY_NAME, newUserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			next.ServeHTTP(w, r)
		}
	}
}

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func createSignedCookie(userID string, secretKey string) (*http.Cookie, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name:     "auth",
		Value:    tokenString,
		Expires:  time.Now().Add(TOKEN_EXP),
		HttpOnly: true,
		Secure:   true,
	}, nil
}

func getUserIDFromCookie(r *http.Request, secretKey string) (string, error) {
	authCookie, err := r.Cookie("auth")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", err
		}
		return "", fmt.Errorf("failed to get auth cookie: %w", err)
	}

	if authCookie.Value == "" {
		return "", errors.New("empty auth cookie")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(authCookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", errors.New("invalid token")
	}

	return claims.UserID, nil
}
