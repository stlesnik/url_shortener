package services

import "github.com/stlesnik/url_shortener/internal/app/repository"

type Repository interface {
	Ping() error
	Save(shortURL string, longURLStr string) error
	Get(shortURL string) (string, bool)
	Close() error
}

type BatchSaver interface {
	SaveBatch(entries []repository.URLPair) error
}
