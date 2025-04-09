package handlers

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"io"
	"net/http"
	"net/url"
)

type Handler struct {
	service *services.URLShortenerService // Вместо прямого доступа к repo и cfg
}

func NewHandler(service *services.URLShortenerService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) getLongURLFromReq(req *http.Request) (string, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", errors.New("error reading body")
	}
	longURLStr := string(body)
	if longURLStr == "" {
		return "", errors.New("didnt get url")
	}
	_, err = url.ParseRequestURI(longURLStr)
	if err != nil {
		errorText := fmt.Sprintf("got incorrect url to shorten: url=%v, err=%v", longURLStr, err.Error())
		return "", errors.New(errorText)
	}
	return longURLStr, nil
}

func (h *Handler) SaveURL(res http.ResponseWriter, req *http.Request) {
	//get long url from body
	longURLStr, err := h.getLongURLFromReq(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	//generate and save short url
	shortURL, err := h.service.CreateShortURL(longURLStr)

	if err == nil {
		//generate response
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		_, err := res.Write([]byte(shortURL))
		if err != nil {
			http.Error(res, "Failed to write short url into response", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(res, "Failed to save short url", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetLongURL(res http.ResponseWriter, req *http.Request) {
	URLHash := chi.URLParam(req, "id")
	longURLStr, err := h.service.GetLongURLFromDB(URLHash)
	if err == nil {
		res.Header().Set("Location", longURLStr)
		res.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(res, "Short url not found", http.StatusBadRequest)
		return
	}
}
