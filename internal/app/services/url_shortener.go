package services

import (
	"errors"
	"fmt"
	"github.com/stlesnik/url_shortener/cmd/config"
)

func PrepareShortURL(urlHash string, cfg *config.Config) string {
	return fmt.Sprintf("%s/%s", cfg.BaseURL, urlHash)

}

type Repository interface {
	Save(shortURL string, longURLStr string) error
	Get(shortURL string) (string, bool)
}

type URLShortenerService struct {
	repo Repository
	cfg  *config.Config
}

func NewURLShortenerService(repo Repository, cfg *config.Config) *URLShortenerService {
	return &URLShortenerService{repo, cfg}
}

func (s *URLShortenerService) CreateShortURL(longURL string) (string, error) {
	urlHash := GenerateShortKey(longURL)
	err := s.repo.Save(urlHash, longURL)
	if err != nil {
		return "", err
	}
	return PrepareShortURL(urlHash, s.cfg), nil
}

func (s *URLShortenerService) GetLongURLFromDB(URLHash string) (string, error) {
	longURL, exists := s.repo.Get(URLHash)
	if !exists {
		return "", errors.New("not found")
	}
	return longURL, nil
}
