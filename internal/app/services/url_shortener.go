package services

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/stlesnik/url_shortener/cmd/config"
	"hash/fnv"
)

type URLShortenerService struct {
	repo Repository
	cfg  *config.Config
}

func NewURLShortenerService(repo Repository, cfg *config.Config) *URLShortenerService {
	return &URLShortenerService{repo, cfg}
}

func (s *URLShortenerService) CreateSavePrepareShortURL(longURL string) (string, string) {
	urlHash, err := s.CreateShortURLHash(longURL)
	if err != nil {
		return "", "Failed to create short url"
	}
	err = s.SaveShortURL(urlHash, longURL)
	if err != nil {
		return "", "Failed to save short url"
	}
	shortURL := s.PrepareShortURL(urlHash)
	return shortURL, ""
}

func (s *URLShortenerService) CreateShortURLHash(longURL string) (string, error) {
	h := fnv.New64a()
	_, err := h.Write([]byte(longURL))
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(h.Sum(nil)), nil
}

func (s *URLShortenerService) SaveShortURL(urlHash, longURL string) error {
	err := s.repo.Save(urlHash, longURL)
	return err
}

func (s *URLShortenerService) PrepareShortURL(urlHash string) string {
	return fmt.Sprintf("%s/%s", s.cfg.BaseURL, urlHash)
}

func (s *URLShortenerService) GetLongURLFromDB(URLHash string) (string, error) {
	longURL, exists := s.repo.Get(URLHash)
	if !exists {
		return "", errors.New("not found")
	}
	return longURL, nil
}
