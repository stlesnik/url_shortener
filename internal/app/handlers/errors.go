package handlers

import (
	"errors"
	"github.com/stlesnik/url_shortener/internal/logger"
	"net/http"
)

var (
	ErrReadingBody = errors.New("error reading body")
	ErrDidntGetURL = errors.New("error getting url")
	ErrInvalidURL  = errors.New("invalid url to shorten")
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
