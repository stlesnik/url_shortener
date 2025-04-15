package handlers

import (
	"encoding/json"
	"errors"
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
	shortURL, errText := h.service.CreateSavePrepareShortURL(longURLStr)
	if errText != "" {
		logger.Sugaarz.Errorw(errText)
		http.Error(res, errText, http.StatusInternalServerError)
		return
	}

	//generate response
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write([]byte(shortURL))
	if err != nil {
		http.Error(res, "Failed to write short url into response", http.StatusInternalServerError)
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

func (h *Handler) APIPrepareShortURL(res http.ResponseWriter, req *http.Request) {
	logger.Sugaarz.Debugw("got APIPrepareShortURL request")
	var apiReq models.APIRequestPrepareShURL
	err := json.NewDecoder(req.Body).Decode(&apiReq)
	if err != nil {
		logger.Sugaarz.Errorw("error decoding body", "err", err)
		http.Error(res, "Failed to decode body", http.StatusInternalServerError)
		return
	}

	shortURL, errText := h.service.CreateSavePrepareShortURL(apiReq.LongURL)
	if errText != "" {
		logger.Sugaarz.Errorw(errText)
		http.Error(res, errText, http.StatusInternalServerError)
		return
	}

	apiResp := models.APIResponsePrepareShURL{
		ShortURL: shortURL,
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(res).Encode(apiResp); err != nil {
		logger.Sugaarz.Errorw("error encoding body", "err", err)
		http.Error(res, "Failed to encode body", http.StatusInternalServerError)
		return
	}
	logger.Sugaarz.Debugw("sent APIPrepareShortURL response")
}
