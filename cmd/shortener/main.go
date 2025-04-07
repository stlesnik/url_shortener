package main

import (
	"github.com/stlesnik/url_shortener/cmd/config"
	"github.com/stlesnik/url_shortener/internal/app/repository"
	"github.com/stlesnik/url_shortener/internal/app/server"
	"log"
)

func main() {
	cfg := config.NewConfig()
	repo := repository.NewInMemoryRepository()
	srv := server.NewServer(repo, cfg)

	log.Printf("Сервер запущен на %s", cfg.ServerAddress)
	err := srv.Start()
	if err != nil {
		panic(err)
	}
}
