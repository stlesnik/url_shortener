package services

import (
	"fmt"
	"net/http"
)

func PrepareShortURL(shortURL string, r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	host := r.Host
	if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}

	return fmt.Sprintf("%s://%s/%s", scheme, host, shortURL)
}
