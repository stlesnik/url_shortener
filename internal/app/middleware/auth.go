package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/stlesnik/url_shortener/internal/config"
	"github.com/stlesnik/url_shortener/internal/logger"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type contextKey string

const (
	TokenExp                 = time.Hour * 24
	UserIDKeyName contextKey = "userID"
)

func WithAuth(cfg *config.Config, createIfNot bool, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserIDFromCookie(r, cfg.AuthSecretKey)

		if err == nil {
			logger.Sugaarz.Infow("Got user id from cookie", "userID", userID)
			ctx := context.WithValue(r.Context(), UserIDKeyName, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else if createIfNot {
			if errors.Is(err, http.ErrNoCookie) {
				logger.Sugaarz.Infow("no token in cookie", "err", err)
			}
			newUserID := uuid.New().String()
			logger.Sugaarz.Infow("No user id in cookie. Created new", "userID", newUserID)
			cookie, err := createSignedCookie(newUserID, cfg.AuthSecretKey)
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, cookie)
			w.Header().Set("Authorization", "Bearer "+cookie.Value)
			ctx := context.WithValue(r.Context(), UserIDKeyName, newUserID)
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
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name:     "Authorization",
		Value:    tokenString,
		Expires:  time.Now().Add(TokenExp),
		HttpOnly: true,
		Secure:   true,
	}, nil
}

func getUserIDFromCookie(r *http.Request, secretKey string) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("failed to get Authorization cookie")
	}

	authToken := strings.Split(auth, " ")
	if len(authToken) != 2 || authToken[0] != "Bearer" {
		return "", fmt.Errorf("invalid Authorization header")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(authToken[1], claims, func(token *jwt.Token) (interface{}, error) {
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
