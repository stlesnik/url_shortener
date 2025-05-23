package services

import (
	"context"
	"github.com/stlesnik/url_shortener/internal/app/models"
	"github.com/stlesnik/url_shortener/internal/app/repository"
)

type Repository interface {
	Ping(ctx context.Context) error
	SaveURL(ctx context.Context, shortURL string, longURLStr string, userID string) (bool, error)
	GetURL(ctx context.Context, shortURL string) (models.GetURLDTO, error)
	Close() error
}

type DBRepository interface {
	Repository
	GetURLList(ctx context.Context, userID string) ([]models.BaseURLDTO, error)
	SaveBatchURL(ctx context.Context, entries []repository.URLPair) error
	DeleteURLList(values []interface{}, placeholders []string) (int64, error)
}
