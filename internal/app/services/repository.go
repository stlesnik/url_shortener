package services

import (
	"context"

	"github.com/stlesnik/url_shortener/internal/app/models"
	"github.com/stlesnik/url_shortener/internal/app/repository"
)

var (
	_ DBStorager = (*repository.DataBase)(nil)
	_ Storager   = (*repository.FileStorage)(nil)
	_ Storager   = (*repository.InMemoryRepository)(nil)
)

// Storager is a URL repository interface.
type Storager interface {
	Ping(ctx context.Context) error
	SaveURL(ctx context.Context, shortURL string, longURLStr string, userID string) (bool, error)
	GetURL(ctx context.Context, shortURL string) (models.GetURLDTO, error)
	Close() error
}

// DBStorager is a DB-specific URL repository interface.
type DBStorager interface {
	Storager
	GetURLList(ctx context.Context, userID string) ([]models.BaseURLDTO, error)
	SaveBatchURL(ctx context.Context, entries []repository.URLPair) error
	DeleteURLList(values []interface{}, placeholders []string) (int64, error)
}
