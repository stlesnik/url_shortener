package main

import (
	"fmt"
	"github.com/stlesnik/url_shortener/internal/app/repository"
	"github.com/stlesnik/url_shortener/internal/app/server"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"github.com/stlesnik/url_shortener/internal/config"
	"github.com/stlesnik/url_shortener/internal/logger"
	"log"
)

func main() {
	// конфиг
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Не получилось создать конфиг: %s", err)
		return
	}

	// логгер
	if logErr := logger.InitLogger(cfg.Environment); logErr != nil {
		panic(fmt.Errorf("logger broke: %w", logErr))
	}
	defer func() {
		if err := logger.Sugaarz.Sync(); err != nil {
			logger.Sugaarz.Errorw("Failed to sync logger", "error", err)
		}
	}()

	var repo services.Repository
	if cfg.FileStoragePath != "" {
		fStorage, err := repository.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			log.Fatalf("ошибка инициализации файлового хранилища: %v", err)
		}
		repo = fStorage
	} else {
		repo = repository.NewInMemoryRepository()
	}
	srv := server.NewServer(repo, cfg)

	log.Printf("Сервер запущен на %s", cfg.ServerAddress)
	err = srv.Start()
	if err != nil {
		log.Fatalf("Не получилось запустить сервер: %s", err)
		return
	}
}
