package services

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/stlesnik/url_shortener/internal/app/models"
	"github.com/stlesnik/url_shortener/internal/app/repository"
	"github.com/stlesnik/url_shortener/internal/config"
	"github.com/stlesnik/url_shortener/internal/logger"
	"hash/fnv"
	"net/url"
)

var (
	ErrServiceSave = errors.New("save error")
)

type URLShortenerService struct {
	repo Repository
	cfg  *config.Config
}

func New(repo Repository, cfg *config.Config) *URLShortenerService {
	return &URLShortenerService{repo, cfg}
}

func (s *URLShortenerService) CreateSavePrepareShortURL(ctx context.Context, longURL string, userID string) (string, bool, string) {
	urlHash, err := s.CreateShortURLHash(longURL)
	if err != nil {
		return "", false, "Failed to create short URL, err: " + err.Error()
	}
	isDouble, err := s.SaveShortURL(ctx, urlHash, longURL, userID)
	if err != nil {
		return "", false, "Failed to save short url, err: " + err.Error()
	}
	shortURL := s.PrepareShortURL(urlHash)
	return shortURL, isDouble, ""
}

func (s *URLShortenerService) CreateShortURLHash(longURL string) (string, error) {
	h := fnv.New64a()
	_, err := h.Write([]byte(longURL))
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(h.Sum(nil)), nil
}

func (s *URLShortenerService) SaveShortURL(ctx context.Context, urlHash, longURL string, userID string) (isDouble bool, err error) {
	isDouble, err = s.repo.SaveURL(ctx, urlHash, longURL, userID)
	return
}

func (s *URLShortenerService) SaveBatchShortURL(ctx context.Context, urlPairList []repository.URLPair) error {
	if bSaver, ok := s.repo.(BatchSaver); ok {
		logger.Sugaarz.Debugw("saving batch urls with BatchSaver")
		err := bSaver.SaveBatchURL(ctx, urlPairList)
		if err != nil {
			return err
		}
	} else {
		logger.Sugaarz.Debugw("saving batch urls ordinary way")
		for _, urlPair := range urlPairList {
			_, err := s.repo.SaveURL(ctx, urlPair.URLHash, urlPair.LongURL, "")
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

func (s *URLShortenerService) GetLongURLFromDB(ctx context.Context, URLHash string) (string, error) {
	longURL, err := s.repo.GetURL(ctx, URLHash)
	return longURL, err
}

func (s *URLShortenerService) GetUserURLs(ctx context.Context, userID string) ([]models.BaseURLResponse, error) {
	if urlListGetter, ok := s.repo.(URLList); ok {
		logger.Sugaarz.Debugw("getting urls for userID")
		urlList, err := urlListGetter.GetURLList(ctx, userID)
		if err != nil {
			return nil, err
		}

		var resp []models.BaseURLResponse
		for _, baseURLObj := range urlList {
			resp = append(resp, models.BaseURLResponse{
				ShortURL:    s.PrepareShortURL(baseURLObj.ShortURLHash),
				OriginalURL: baseURLObj.OriginalURL,
			})
		}
		return resp, nil
	} else {
		return nil, errors.New("not implemented error")
	}
}

func (s *URLShortenerService) PingDB(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
