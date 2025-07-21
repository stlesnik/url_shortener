package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/stlesnik/url_shortener/internal/app/models"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"github.com/stlesnik/url_shortener/internal/logger"
)

type Handler struct {
	service *services.URLShortenerService // Вместо прямого доступа к repo и cfg
}

func New(service *services.URLShortenerService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) SaveURL(res http.ResponseWriter, req *http.Request) {
	//get user id
	userID, err := h.service.GetUserID(req)
	if errors.Is(err, services.ErrNoUserID) {
		logger.Sugaarz.Warn(err.Error())
		WriteError(res, err.Error(), http.StatusUnauthorized, false)
		return
	}
	if errors.Is(err, services.ErrConvertingUserID) {
		logger.Sugaarz.Warn(err.Error())
		WriteError(res, err.Error(), http.StatusInternalServerError, true)
		return
	}

	//get long url from body
	longURLStr, err := h.service.GetLongURLFromReq(req)
	if err != nil {
		WriteError(res, err.Error(), http.StatusBadRequest, true)
		return
	}
	//generate and save short url
	shortURL, isDouble, err := h.service.GenerateShortURL(req.Context(), longURLStr, userID)
	if err != nil {
		logger.Sugaarz.Errorw("error while generating short URL", "err", err)
		WriteError(res, err.Error(), http.StatusInternalServerError, true)
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
	URLHash := h.service.GetURLHash(req)
	urlDTO, err := h.service.GetLongURLFromDB(req.Context(), URLHash)

	if err != nil {
		WriteError(res, "Short url not found", http.StatusBadRequest, false)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if !urlDTO.IsDeleted {
		res.Header().Set("Location", urlDTO.OriginalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	res.WriteHeader(http.StatusGone)
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

	shortURL, isDouble, err := h.service.GenerateShortURL(req.Context(), apiReq.LongURL, "")
	if err != nil {
		logger.Sugaarz.Errorw("error while generating short URL", "err", err)
		WriteError(res, err.Error(), http.StatusInternalServerError, true)
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
	apiBatchResp, batch, validationErrors, err := h.service.PrepareBatch(apiBatchReq)
	if err != nil {
		logger.Sugaarz.Errorw("failed to create short URL", "err", err)
		WriteError(res, "failed to process api batch: "+err.Error(), http.StatusInternalServerError, true)
	}
	//save batch
	txErr := h.service.SaveBatchShortURL(req.Context(), batch)
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

func (h *Handler) APIGetUserURLs(res http.ResponseWriter, req *http.Request) {
	logger.Sugaarz.Debugw("got APIGetUserURLs response")
	userID, err := h.service.GetUserID(req)
	if errors.Is(err, services.ErrNoUserID) {
		logger.Sugaarz.Warn(err.Error())
		WriteError(res, err.Error(), http.StatusUnauthorized, false)
		return
	}
	if errors.Is(err, services.ErrConvertingUserID) {
		logger.Sugaarz.Warn(err.Error())
		WriteError(res, err.Error(), http.StatusInternalServerError, true)
		return
	}
	var urlsResponseObj []models.BaseURLResponse
	urlsResponseObj, URLErr := h.service.GetUserURLs(req.Context(), userID)
	if URLErr != nil {
		logger.Sugaarz.Errorw("error getting users urls", "err", URLErr)
		WriteError(res, "error getting users urls", http.StatusNoContent, false)
		return
	}
	if len(urlsResponseObj) == 0 {
		res.WriteHeader(http.StatusNoContent)
		logger.Sugaarz.Debugw("sent APIGetUserURLs response no content")
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(urlsResponseObj); err != nil {
		logger.Sugaarz.Errorw("error encoding body", "err", err)
		WriteError(res, "failed to encode body", http.StatusInternalServerError, true)
		return
	}
}

func (h *Handler) APIDeleteUserURLs(res http.ResponseWriter, req *http.Request) {
	logger.Sugaarz.Debugw("got APIDeleteUserURLs response")
	userID, err := h.service.GetUserID(req)
	if errors.Is(err, services.ErrNoUserID) {
		logger.Sugaarz.Warn(err.Error())
		WriteError(res, err.Error(), http.StatusUnauthorized, false)
		return
	}
	if errors.Is(err, services.ErrConvertingUserID) {
		logger.Sugaarz.Warn(err.Error())
		WriteError(res, err.Error(), http.StatusInternalServerError, true)
		return
	}

	var urlHashes []string
	err = json.NewDecoder(req.Body).Decode(&urlHashes)
	if err != nil {
		logger.Sugaarz.Errorw("error decoding body", "err", err)
		WriteError(res, "Failed to decode body", http.StatusInternalServerError, true)
		return
	}
	go h.service.SendDeleteTasks(userID, urlHashes)

	res.WriteHeader(http.StatusAccepted)
	logger.Sugaarz.Debugw("sent APIDeleteUserURLs response")
}

func (h *Handler) PingDB(res http.ResponseWriter, req *http.Request) {
	logger.Sugaarz.Debugw("got PingDB request")
	err := h.service.PingDB(req.Context())
	if err == nil {
		res.WriteHeader(http.StatusOK)
	} else {
		WriteError(res, err.Error(), http.StatusInternalServerError, true)
		return
	}
}
