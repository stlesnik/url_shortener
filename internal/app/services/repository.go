package services

import (
	"context"

	"github.com/stlesnik/url_shortener/internal/app/models"
	"github.com/stlesnik/url_shortener/internal/app/repository"
)

var (
	_ IDBRepository = (*repository.DataBase)(nil)
	_ IRepository   = (*repository.FileStorage)(nil)
	_ IRepository   = (*repository.InMemoryRepository)(nil)
)

// IRepository defines the interface for a basic URL repository.
type IRepository interface {
	Ping(ctx context.Context) error
	SaveURL(ctx context.Context, shortURL string, longURLStr string, userID string) (bool, error)
	GetURL(ctx context.Context, shortURL string) (models.GetURLDTO, error)
	Close() error
}

// IDBRepository extends IRepository with db-specific methods.
type IDBRepository interface {
	IRepository
	GetURLList(ctx context.Context, userID string) ([]models.BaseURLDTO, error)
	SaveBatchURL(ctx context.Context, entries []repository.URLPair) error
	DeleteURLList(values []interface{}, placeholders []string) (int64, error)
}
