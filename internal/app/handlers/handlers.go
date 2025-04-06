package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"net/http"
)

type Handler struct {
	service *services.UrlShortenerService // Вместо прямого доступа к repo и cfg
}

func NewHandler(service *services.UrlShortenerService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SaveURL(res http.ResponseWriter, req *http.Request) {
	//get long url from body
	longURLStr, err := services.GetLongURLFromReq(req)
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
		res.Write([]byte(shortURL))
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
