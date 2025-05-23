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
	"time"
)

const (
	bufferSize           = 5
	deleteTickerInterval = 1 * time.Second
)

var (
	ErrServiceSave = errors.New("save error")
)

type URLShortenerService struct {
	repo          Repository
	cfg           *config.Config
	deleteCh      chan models.DeleteTask
	daemonsDoneCh chan struct{}
}

func New(repo Repository, cfg *config.Config, daemonsDoneCh chan struct{}) *URLShortenerService {
	s := &URLShortenerService{repo, cfg, make(chan models.DeleteTask, bufferSize), daemonsDoneCh}
	return s.init()
}
func (s *URLShortenerService) init() *URLShortenerService {
	if _, ok := s.repo.(DBRepository); ok {
		logger.Sugaarz.Debugw("starting DeleteUrls goroutine")
		go s.DeleteUrls()
	}
	return s
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
	if rep, ok := s.repo.(DBRepository); ok {
		logger.Sugaarz.Debugw("saving batch urls with BatchSaver")
		err := rep.SaveBatchURL(ctx, urlPairList)
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

func (s *URLShortenerService) GetLongURLFromDB(ctx context.Context, URLHash string) (models.GetURLDTO, error) {
	urlDTO, err := s.repo.GetURL(ctx, URLHash)
	return urlDTO, err
}

func (s *URLShortenerService) GetUserURLs(ctx context.Context, userID string) ([]models.BaseURLResponse, error) {
	if rep, ok := s.repo.(DBRepository); ok {
		logger.Sugaarz.Debugw("getting urls for userID")
		urlList, err := rep.GetURLList(ctx, userID)
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

func (s *URLShortenerService) GenerateDeleteTasks(userID string, urlHashes []string) {
	if _, ok := s.repo.(DBRepository); ok {
		for _, urlHash := range urlHashes {
			s.deleteCh <- models.DeleteTask{UserID: userID, URLHash: urlHash}
		}
		logger.Sugaarz.Debug("Created ", len(urlHashes), " delete tasks")
	} else {
		logger.Sugaarz.Error("not implemented error")
	}

}

func (s *URLShortenerService) DeleteUrls() {
	ticker := time.NewTicker(deleteTickerInterval)

	var (
		values       []interface{}
		placeholders []string
	)
	plInd := 1
	do := func() {
		logger.Sugaarz.Debugw("deleting urls for userID")
		rowsAffected, err := s.repo.(DBRepository).DeleteURLList(values, placeholders)
		if err != nil {
			logger.Sugaarz.Error(err)
		} else {
			logger.Sugaarz.Debug(rowsAffected, "rows were updated on delete")
		}
		values = nil
		placeholders = nil
		plInd = 1
	}

loop:
	for {
		select {
		case <-s.daemonsDoneCh:
			if len(values) == 0 {
				continue
			}
			do()
			break loop
		case task := <-s.deleteCh:
			values = append(values, task.UserID, task.URLHash)
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", plInd, plInd+1))
			plInd = plInd + 2

		case <-ticker.C:
			if len(values) == 0 {
				continue
			}
			do()
		}
	}
}

func (s *URLShortenerService) PingDB(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
