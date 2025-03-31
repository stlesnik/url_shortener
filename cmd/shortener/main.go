package main

import (
	"github.com/stlesnik/url_shortener/internal/app/server"
	"github.com/stlesnik/url_shortener/internal/app/storage"
)

func main() {
	repo := storage.NewInMemoryStorage()
	srv := server.NewServer(repo, ":8080")
	err := srv.Start()
	if err != nil {
		panic(err)
	}
}
