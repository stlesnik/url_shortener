package main

import (
	"github.com/stlesnik/url_shortener/cmd/config"
	"github.com/stlesnik/url_shortener/internal/app/server"
	"github.com/stlesnik/url_shortener/internal/app/storage"
	"log"
)

func main() {
	cfg := config.NewConfig()
	repo := storage.NewInMemoryStorage()
	srv := server.NewServer(repo, cfg)

	log.Printf("Сервер запущен на %s", cfg.ServerAddress)
	err := srv.Start()
	if err != nil {
		panic(err)
	}
}
