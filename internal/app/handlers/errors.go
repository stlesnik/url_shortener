package handlers

import (
	"github.com/stlesnik/url_shortener/cmd/logger"
	"net/http"
)

func WriteError(w http.ResponseWriter, msg string, code int) {
	logger.Sugaarz.Errorw(msg, "code", code)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(msg))
}
