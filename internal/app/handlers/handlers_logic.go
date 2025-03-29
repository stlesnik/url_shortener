package handlers

import (
	"github.com/stlesnik/url_shortener/internal/app/utils"
	"io"
	"net/http"
	"net/url"
)

func processPostRequest(res http.ResponseWriter, req *http.Request) {
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

	shortURL := utils.GenerateShortKey(longURLStr)
	urlMap[shortURL] = longURLStr

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(utils.PrepareShortURL(shortURL, req)))
}

func processGetRequest(res http.ResponseWriter, id string) {
	longURLStr, exists := urlMap[id]
	if exists {
		res.Header().Set("Location", longURLStr)
		res.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(res, "Short url not found", http.StatusBadRequest)
	}
}
