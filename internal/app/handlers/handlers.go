package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/stlesnik/url_shortener/cmd/config"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"github.com/stlesnik/url_shortener/internal/app/storage"
	"net/http"
)

type Handler struct {
	repo storage.Repository
	cfg  *config.Config
}

func NewHandler(repo storage.Repository, cfg *config.Config) *Handler {
	return &Handler{repo: repo, cfg: cfg}
}

func (h *Handler) SaveURL(res http.ResponseWriter, req *http.Request) {
	//get long url from body
	longURLStr, err := services.GetLongURL(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}
	//generate short url
	urlHash := services.GenerateShortKey(longURLStr)

	//save short url
	err = h.repo.Save(urlHash, longURLStr)

	if err == nil {
		//generate response
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(services.PrepareShortURL(urlHash, h.cfg)))
	} else {
		http.Error(res, "Failed to save short url", http.StatusInternalServerError)
	}
}

func (h *Handler) GetLongURL(res http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	longURLStr, exists := h.repo.Get(id)
	if exists {
		res.Header().Set("Location", longURLStr)
		res.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(res, "Short url not found", http.StatusBadRequest)
	}
}
