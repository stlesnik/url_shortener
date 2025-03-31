package handlers

import (
	"github.com/stlesnik/url_shortener/internal/app/services"
	"github.com/stlesnik/url_shortener/internal/app/storage"
	"io"
	"net/http"
	"net/url"
)

type Handler struct {
	repo storage.Repository
}

func NewHandler(repo storage.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) MainHandler(res http.ResponseWriter, req *http.Request) {
	id, method, err := services.ProcessRequest(res, req)
	if err != nil {
		return
	}
	if method == http.MethodPost {
		h.processPostRequest(res, req)
	} else {
		h.processGetRequest(res, id)
	}
}

func (h *Handler) processPostRequest(res http.ResponseWriter, req *http.Request) {
	longURL, err := io.ReadAll(req.Body)
	longURLStr := string(longURL)
	if err != nil {
		http.Error(res, "Error reading body", http.StatusBadRequest)
		return
	}
	_, err = url.ParseRequestURI(longURLStr)
	if err != nil {
		http.Error(res, "Got incorrect url to shorten", http.StatusBadRequest)
		return
	}

	shortURL := services.GenerateShortKey(longURLStr)
	err = h.repo.Save(shortURL, longURLStr)
	if err == nil {
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(services.PrepareShortURL(shortURL, req)))
	} else {
		http.Error(res, "Failed to save short url", http.StatusInternalServerError)
	}
}

func (h *Handler) processGetRequest(res http.ResponseWriter, id string) {
	longURLStr, exists := h.repo.Get(id)
	if exists {
		res.Header().Set("Location", longURLStr)
		res.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(res, "Short url not found", http.StatusBadRequest)
	}
}
