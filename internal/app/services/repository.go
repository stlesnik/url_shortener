package services

import (
	"context"
	"github.com/stlesnik/url_shortener/internal/app/repository"
)

type Repository interface {
	Ping(ctx context.Context) error
	Save(ctx context.Context, shortURL string, longURLStr string) (bool, error)
	Get(ctx context.Context, shortURL string) (string, error)
	Close() error
}

type BatchSaver interface {
	SaveBatch(ctx context.Context, entries []repository.URLPair) error
}
