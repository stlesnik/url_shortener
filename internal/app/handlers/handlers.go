package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stlesnik/url_shortener/cmd/logger"
	"github.com/stlesnik/url_shortener/internal/app/models"
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
		return "", ErrReadingBody
	}
	longURLStr := string(body)
	if longURLStr == "" {
		return "", ErrDidntGetURL
	}
	_, err = url.ParseRequestURI(longURLStr)
	if err != nil {
		return "", fmt.Errorf("got incorrect url to shorten: url=%v, err=%v: %w", longURLStr, err, ErrInvalidURL)
	}
	return longURLStr, nil
}

func (h *Handler) SaveURL(res http.ResponseWriter, req *http.Request) {
	//get long url from body
	longURLStr, err := h.getLongURLFromReq(req)
	if err != nil {
		WriteError(res, err.Error(), http.StatusBadRequest)
		return
	}
	//generate and save short url
	shortURL, errText := h.service.CreateSavePrepareShortURL(longURLStr)
	if errText != "" {
		logger.Sugaarz.Errorw(errText)
		WriteError(res, errText, http.StatusInternalServerError)
		return
	}

	//generate response
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write([]byte(shortURL))
	if err != nil {
		WriteError(res, "Failed to write short url into response", http.StatusInternalServerError)
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
		WriteError(res, "Short url not found", http.StatusBadRequest)
		return
	}
}

func (h *Handler) APIPrepareShortURL(res http.ResponseWriter, req *http.Request) {
	logger.Sugaarz.Debugw("got APIPrepareShortURL request")
	var apiReq models.APIRequestPrepareShURL
	err := json.NewDecoder(req.Body).Decode(&apiReq)
	if err != nil {
		logger.Sugaarz.Errorw("error decoding body", "err", err)
		WriteError(res, "Failed to decode body", http.StatusInternalServerError)
		return
	}

	shortURL, errText := h.service.CreateSavePrepareShortURL(apiReq.LongURL)
	if errText != "" {
		logger.Sugaarz.Errorw(errText)
		WriteError(res, errText, http.StatusInternalServerError)
		return
	}

	apiResp := models.APIResponsePrepareShURL{
		ShortURL: shortURL,
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(res).Encode(apiResp); err != nil {
		logger.Sugaarz.Errorw("error encoding body", "err", err)
		WriteError(res, "Failed to encode body", http.StatusInternalServerError)
		return
	}
	logger.Sugaarz.Debugw("sent APIPrepareShortURL response")
}
