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
		log.Fatalf("Не получилось обработать конфиг: %s", err)
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

	// db
	var repo services.Repository

	switch {
	case cfg.DatabaseDSN != "":
		{
			db, dbErr := repository.NewDataBase(cfg.DatabaseDSN)
			if dbErr != nil {
				panic(fmt.Errorf("не получилось подключиться к бд: %w", dbErr))
			}
			pingErr := db.Ping()
			if pingErr != nil {
				panic(fmt.Errorf("не получилось пингануть бд: %w", pingErr))
			}
			repo = db
		}
	case cfg.FileStoragePath != "":
		{
			fStorage, err := repository.NewFileStorage(cfg.FileStoragePath)
			if err != nil {
				log.Fatalf("ошибка инициализации файлового хранилища: %v", err)
			}
			repo = fStorage
		}
	default:
		{
			repo = repository.NewInMemoryRepository()
		}
	}
	defer func() {
		if err := repo.Close(); err != nil {
			logger.Sugaarz.Errorw("Failed to close repository", "error", err)
		}
	}()

	srv := server.NewServer(repo, cfg)

	log.Printf("Сервер запущен на %s", cfg.ServerAddress)
	err = srv.Start()
	if err != nil {
		log.Fatalf("Не получилось запустить сервер: %s", err)
		return
	}
}
