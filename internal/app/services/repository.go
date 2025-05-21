package services

import (
	"context"
	"github.com/stlesnik/url_shortener/internal/app/models"
	"github.com/stlesnik/url_shortener/internal/app/repository"
)

type Repository interface {
	Ping(ctx context.Context) error
	SaveURL(ctx context.Context, shortURL string, longURLStr string, userID string) (bool, error)
	GetURL(ctx context.Context, shortURL string) (string, error)
	Close() error
}

type URLList interface {
	GetURLList(ctx context.Context, userID string) ([]models.BaseURLResponse, error)
}
type BatchSaver interface {
	SaveBatchURL(ctx context.Context, entries []repository.URLPair) error
}
