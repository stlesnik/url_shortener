package handlers

import (
	"net/http"

	"github.com/stlesnik/url_shortener/internal/logger"
)

func WriteError(w http.ResponseWriter, msg string, code int, trace bool) {
	if trace {
		logger.Sugaarz.Errorw(msg, "code", code)
	} else {
		logger.Sugaarz.Infow(msg, "code", code)
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(msg))
}
