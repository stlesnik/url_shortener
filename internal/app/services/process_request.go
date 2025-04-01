package services

import (
	"errors"
	"net/http"
	"strings"
)

func ProcessRequest(res http.ResponseWriter, req *http.Request) (string, string, error) {
	id := strings.TrimPrefix(req.URL.Path, "/")
	if strings.Contains(id, "/") {
		http.Error(res, "incorrect url", http.StatusBadRequest)
		return "", "", errors.New("incorrect url")
	}
	method := req.Method
	if method != http.MethodPost && method != http.MethodGet {
		http.Error(res, "incorrect method: only GET and POST allowed", http.StatusBadRequest)
		return "", "", errors.New("incorrect method: only GET and POST allowed")
	}
	return id, method, nil
}
