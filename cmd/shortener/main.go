package main

import (
	"github.com/stlesnik/url_shortener/internal/app/handlers"
	"net/http"
)

func main() {
	mux := handlers.Init()
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
