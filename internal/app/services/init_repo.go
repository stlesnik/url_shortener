package services

import (
	"context"
	"fmt"

	"github.com/stlesnik/url_shortener/internal/app/repository"
	"github.com/stlesnik/url_shortener/internal/config"
	"github.com/stlesnik/url_shortener/internal/migrator"
)

// db

// NewRepository инициализирует и возвращает подходящий репозиторий в зависимости от конфигурации.
func NewRepository(cfg *config.Config) (Storager, error) {
	var repo Storager

	switch {
	case cfg.DatabaseDSN != "":
		{
			migrator.Run(cfg.DatabaseDSN)

			db, dbErr := repository.NewDataBase(cfg.DatabaseDSN)
			if dbErr != nil {
				return nil, fmt.Errorf("не получилось подключиться к бд: %w", dbErr)
			}
			pingErr := db.Ping(context.Background())
			if pingErr != nil {
				return nil, fmt.Errorf("не получилось пингануть бд: %w", pingErr)
			}
			repo = db
		}
	case cfg.FileStoragePath != "":
		{
			fStorage, err := repository.NewFileStorage(cfg.FileStoragePath)
			if err != nil {
				return nil, fmt.Errorf("ошибка инициализации файлового хранилища: %v", err)
			}
			repo = fStorage
		}
	default:
		{
			repo = repository.NewInMemoryRepository()
		}
	}
	return repo, nil
}
