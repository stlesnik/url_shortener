package handlers

import (
	"github.com/stlesnik/url_shortener/internal/app/utils"
	"net/http"
)

var urlMap = make(map[string]string)

func mainHandler(res http.ResponseWriter, req *http.Request) {
	id, method, err := utils.ProcessRequest(res, req)
	if err != nil {
		return
	}
	if method == http.MethodPost {
		processPostRequest(res, req)
	} else {
		processGetRequest(res, id)
	}
}

func Init() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", mainHandler)
	return mux
}
