package services

import (
	"errors"
	"fmt"
	"github.com/stlesnik/url_shortener/cmd/config"
	"github.com/stlesnik/url_shortener/internal/app/storage"
)

func PrepareShortURL(urlHash string, cfg *config.Config) string {
	return fmt.Sprintf("%s/%s", cfg.BaseURL, urlHash)

}

type UrlShortenerService struct {
	repo storage.Repository
	cfg  *config.Config
}

func NewUrlShortenerService(repo storage.Repository, cfg *config.Config) *UrlShortenerService {
	return &UrlShortenerService{repo, cfg}
}

func (s *UrlShortenerService) CreateShortURL(longURL string) (string, error) {
	urlHash := GenerateShortKey(longURL)
	err := s.repo.Save(urlHash, longURL)
	if err != nil {
		return "", err
	}
	return PrepareShortURL(urlHash, s.cfg), nil
}

func (s *UrlShortenerService) GetLongURLFromDB(URLHash string) (string, error) {
	longURL, exists := s.repo.Get(URLHash)
	if !exists {
		return "", errors.New("not found")
	}
	return longURL, nil
}
