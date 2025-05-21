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
		if err := logger.Sugaarz.Sync(); err != nil {
			logger.Sugaarz.Errorw("failed to sync logger", "error", err)
		}
	}()

	repo, err := services.NewRepository(cfg)
	if err != nil {
		logger.Sugaarz.Errorw("failed to create repository", "error", err)
		return
	}
	defer func() {
		if err := repo.Close(); err != nil {
			logger.Sugaarz.Errorw("Failed to close repository", "error", err)
		}
	}()

	srv := server.New(repo, cfg)

	log.Printf("Сервер запущен на %s", cfg.ServerAddress)
	err = srv.Start()
	if err != nil {
		log.Fatalf("Не получилось запустить сервер: %s", err)
		return
	}
}
