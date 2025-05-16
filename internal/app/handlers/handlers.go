package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stlesnik/url_shortener/internal/app/models"
	"github.com/stlesnik/url_shortener/internal/app/repository"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"github.com/stlesnik/url_shortener/internal/logger"
	"io"
	"net/http"
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
	err = h.service.ValidateURL(longURLStr)
	if err != nil {
		return "", fmt.Errorf("got incorrect url to shorten: url=%v, err=%v: %w", longURLStr, err, ErrInvalidURL)
	}
	return longURLStr, nil
}

func (h *Handler) SaveURL(res http.ResponseWriter, req *http.Request) {
	//get long url from body
	longURLStr, err := h.getLongURLFromReq(req)
	if err != nil {
		WriteError(res, err.Error(), http.StatusBadRequest, true)
		return
	}
	//generate and save short url
	shortURL, isDouble, errText := h.service.CreateSavePrepareShortURL(longURLStr)
	if errText != "" {
		logger.Sugaarz.Errorw(errText)
		WriteError(res, errText, http.StatusInternalServerError, true)
		return
	}

	//generate response
	res.Header().Set("Content-Type", "text/plain")
	if isDouble {
		res.WriteHeader(http.StatusConflict)

	} else {
		res.WriteHeader(http.StatusCreated)

	}
	_, err = res.Write([]byte(shortURL))
	if err != nil {
		WriteError(res, "Failed to write short url into response", http.StatusInternalServerError, true)
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
		WriteError(res, "Short url not found", http.StatusBadRequest, false)
		return
	}
}

func (h *Handler) APIPrepareShortURL(res http.ResponseWriter, req *http.Request) {
	logger.Sugaarz.Debugw("got APIPrepareShortURL request")
	var apiReq models.APIRequestPrepareShURL
	err := json.NewDecoder(req.Body).Decode(&apiReq)
	if err != nil {
		logger.Sugaarz.Errorw("error decoding body", "err", err)
		WriteError(res, "Failed to decode body", http.StatusInternalServerError, true)
		return
	}
	validateErr := h.service.ValidateURL(apiReq.LongURL)
	if validateErr != nil {
		logger.Sugaarz.Errorw("got incorrect url to shorten: "+apiReq.LongURL, "err", err)
		WriteError(res, "got incorrect url to shorten: "+apiReq.LongURL, http.StatusInternalServerError, true)
		return
	}

	shortURL, isDouble, errText := h.service.CreateSavePrepareShortURL(apiReq.LongURL)
	if errText != "" {
		logger.Sugaarz.Errorw(errText)
		WriteError(res, errText, http.StatusInternalServerError, true)
		return
	}

	apiResp := models.APIResponsePrepareShURL{
		ShortURL: shortURL,
	}
	res.Header().Set("Content-Type", "application/json")
	if isDouble {
		res.WriteHeader(http.StatusConflict)

	} else {
		res.WriteHeader(http.StatusCreated)

	}
	if err := json.NewEncoder(res).Encode(apiResp); err != nil {
		logger.Sugaarz.Errorw("error encoding body", "err", err)
		WriteError(res, "Failed to encode body", http.StatusInternalServerError, true)
		return
	}
	logger.Sugaarz.Debugw("sent APIPrepareShortURL response")
}

func (h *Handler) APIPrepareBatchShortURL(res http.ResponseWriter, req *http.Request) {
	//process request
	logger.Sugaarz.Debugw("got APISaveBatchURL request")
	var apiBatchReq []models.APIRequestPrepareBatchShURL
	err := json.NewDecoder(req.Body).Decode(&apiBatchReq)
	if err != nil {
		logger.Sugaarz.Errorw("error decoding body", "err", err)
		WriteError(res, "Failed to decode body", http.StatusInternalServerError, true)
		return
	}
	//prepare db and response
	var (
		apiBatchResp     []models.APIResponsePrepareBatchShURL
		batch            []repository.URLPair
		validationErrors []error
	)
	for _, obj := range apiBatchReq {
		validateErr := h.service.ValidateURL(obj.LongURL)
		if validateErr != nil {
			logger.Sugaarz.Errorw("got incorrect url to shorten: "+obj.LongURL, "err", validateErr)
			validationErrors = append(validationErrors, validateErr)
		} else {
			urlHash, err := h.service.CreateShortURLHash(obj.LongURL)
			if err != nil {
				logger.Sugaarz.Errorw("Failed to create short URL", "err", err)
				WriteError(res, "Failed to create short URL,", http.StatusInternalServerError, true)
				return
			}
			batch = append(batch, repository.URLPair{URLHash: urlHash, LongURL: obj.LongURL})
			apiBatchResp = append(apiBatchResp, models.APIResponsePrepareBatchShURL{
				CorrelationID: obj.CorrelationID, ShortURL: h.service.PrepareShortURL(urlHash)})
		}
	}
	//save batch
	txErr := h.service.SaveBatchShortURL(batch)
	if txErr != nil {
		logger.Sugaarz.Errorw("error while saving batch", "err", txErr)
		WriteError(res, "error while saving batch", http.StatusInternalServerError, true)
		return
	}

	//create response
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(res).Encode(apiBatchResp); err != nil {
		logger.Sugaarz.Errorw("error encoding body", "err", err)
		WriteError(res, "Failed to encode body", http.StatusInternalServerError, true)
		return
	}
	if len(validationErrors) > 0 {
		logger.Sugaarz.Errorf("got %v errors while processing request: %w", len(validationErrors), errors.Join(validationErrors...))
	}
	logger.Sugaarz.Debugw("sent APISaveBatchURL response")
}

func (h *Handler) PingDB(res http.ResponseWriter, _ *http.Request) {
	logger.Sugaarz.Debugw("got PingDB request")
	err := h.service.PingDB()
	if err == nil {
		res.WriteHeader(http.StatusOK)
	} else {
		WriteError(res, err.Error(), http.StatusInternalServerError, true)
		return
	}
}
