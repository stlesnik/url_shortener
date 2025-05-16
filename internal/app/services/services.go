package services

import (
	"encoding/base64"
	"fmt"
	"github.com/stlesnik/url_shortener/internal/app/repository"
	"github.com/stlesnik/url_shortener/internal/config"
	"github.com/stlesnik/url_shortener/internal/logger"
	"hash/fnv"
	"net/url"
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
		return "", "Failed to create short URL, err: " + err.Error()
	}
	err = s.SaveShortURL(urlHash, longURL)
	if err != nil {
		return "", "Failed to save short url, err: " + err.Error()
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

func (s *URLShortenerService) SaveBatchShortURL(urlPairList []repository.URLPair) error {
	if bSaver, ok := s.repo.(BatchSaver); ok {
		logger.Sugaarz.Debugw("saving batch urls with BatchSaver")
		err := bSaver.SaveBatch(urlPairList)
		if err != nil {
			return err
		}
	} else {
		logger.Sugaarz.Debugw("saving batch urls ordinary way")
		for _, urlPair := range urlPairList {
			err := s.repo.Save(urlPair.URLHash, urlPair.LongURL)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *URLShortenerService) ValidateURL(longURL string) error {
	_, err := url.ParseRequestURI(longURL)
	if err != nil {
		return fmt.Errorf("got incorrect url to shorten: url=%v, err= %w", longURL, err)
	}
	return nil
}

func (s *URLShortenerService) PrepareShortURL(urlHash string) string {
	return fmt.Sprintf("%s/%s", s.cfg.BaseURL, urlHash)
}

func (s *URLShortenerService) GetLongURLFromDB(URLHash string) (string, error) {
	longURL, exists := s.repo.Get(URLHash)
	if !exists {
		return "", ErrURLNotFound
	}
	return longURL, nil
}

func (s *URLShortenerService) PingDB() error {
	return s.repo.Ping()
}
