package main

import (
	"fmt"
	"github.com/stlesnik/url_shortener/internal/app/server"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"github.com/stlesnik/url_shortener/internal/config"
	"github.com/stlesnik/url_shortener/internal/logger"
	"log"
)

func main() {
	daemonsDoneCh := make(chan struct{})
	defer close(daemonsDoneCh)
	// конфиг
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Не получилось обработать конфиг: %s", err)
		return
	}

	// логгер
	if logErr := logger.InitLogger(cfg.Environment); logErr != nil {
		panic(fmt.Errorf("logger broke: %w", logErr))
	}
	defer func() {
		if syncErr := logger.Sugaarz.Sync(); syncErr != nil {
			logger.Sugaarz.Errorw("failed to sync logger", "error", syncErr)
		}
	}()

	repo, err := services.NewRepository(cfg)
	if err != nil {
		logger.Sugaarz.Errorw("failed to create repository", "error", err)
		return
	}
	defer func() {
		if closeErr := repo.Close(); closeErr != nil {
			logger.Sugaarz.Errorw("Failed to close repository", "error", closeErr)
		}
	}()

	srv, err := server.New(repo, cfg, daemonsDoneCh)
	if err != nil {
		logger.Sugaarz.Errorw("failed to create server", "error", err)
		return
	}

	log.Printf("Сервер запущен на %s", cfg.ServerAddress)
	err = srv.Start()
	if err != nil {
		log.Fatalf("Не получилось запустить сервер: %s", err)
		return
	}
}
