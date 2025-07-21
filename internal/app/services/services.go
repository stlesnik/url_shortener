package services

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/stlesnik/url_shortener/internal/app/middleware"
	"github.com/stlesnik/url_shortener/internal/app/models"
	"github.com/stlesnik/url_shortener/internal/app/repository"
	"github.com/stlesnik/url_shortener/internal/config"
	"github.com/stlesnik/url_shortener/internal/logger"
)

// bufferSize is the size of the delete task channel buffer.
const bufferSize = 10

// deleteTickerInterval is the interval for the delete ticker.
const deleteTickerInterval = 100 * time.Millisecond

// DeleteBatchSize is the batch size for deleting URLs.
const DeleteBatchSize = 100

// ErrServiceSave is returned when there is a service save error.
var ErrServiceSave = errors.New("save error")

// URLShortenerService provides business logic for URL shortening operations.
type URLShortenerService struct {
	repo            Storager
	cfg             *config.Config
	deleteCh        chan models.DeleteTask
	daemonsDoneCh   chan struct{}
	deleteSemaphore chan struct{}
}

// New creates a new URLShortenerService with the given repository and config.
func New(repo Storager, cfg *config.Config) *URLShortenerService {
	return &URLShortenerService{
		repo: repo,
		cfg:  cfg,
	}
}

// InitDeleteDaemon initializes the delete goroutine.
func (s *URLShortenerService) InitDeleteDaemon(daemonsDoneCh chan struct{}) {
	s.deleteCh = make(chan models.DeleteTask, bufferSize)
	s.daemonsDoneCh = daemonsDoneCh
	s.deleteSemaphore = make(chan struct{}, 5)

	if _, ok := s.repo.(DBStorager); ok {
		logger.Sugaarz.Debugw("starting DeleteUrls goroutine")
		go s.DeleteUrls()
	}
}

// GenerateShortURL generates a short URL for the given long URL and user ID.
func (s *URLShortenerService) GenerateShortURL(ctx context.Context, longURL, userID string) (string, bool, error) {
	urlHash, err := s.CreateShortURLHash(longURL)
	if err != nil {
		return "", false, fmt.Errorf("failed to create short URL, err: %w", err)
	}
	isDouble, err := s.SaveShortURL(ctx, urlHash, longURL, userID)
	if err != nil {
		return "", false, fmt.Errorf("failed to save short url, err: %w", err)
	}
	shortURL := s.PrepareShortURL(urlHash)
	return shortURL, isDouble, nil
}

// CreateShortURLHash creates a hash for the given long URL.
func (s *URLShortenerService) CreateShortURLHash(longURL string) (string, error) {
	h := fnv.New64a()
	_, err := h.Write([]byte(longURL))
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(h.Sum(nil)), nil
}

// SaveShortURL saves a short URL mapping to the repository.
func (s *URLShortenerService) SaveShortURL(ctx context.Context, urlHash, longURL string, userID string) (bool, error) {
	isDouble, err := s.repo.SaveURL(ctx, urlHash, longURL, userID)
	return isDouble, err
}

// SaveBatchShortURL saves a batch of URL pairs to the repository.
func (s *URLShortenerService) SaveBatchShortURL(ctx context.Context, urlPairList []repository.URLPair) error {
	if rep, ok := s.repo.(DBStorager); ok {
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

// ValidateURL validates the format of a long URL.
func (s *URLShortenerService) ValidateURL(longURL string) error {
	_, err := url.ParseRequestURI(longURL)
	if err != nil {
		return fmt.Errorf("got incorrect url to shorten: url=%v, err= %w", longURL, err)
	}
	return nil
}

// PrepareShortURL prepares the short URL string from a hash.
func (s *URLShortenerService) PrepareShortURL(urlHash string) string {
	return fmt.Sprintf("%s/%s", s.cfg.BaseURL, urlHash)
}

// GetLongURLFromDB retrieves the original URL from the repository by its hash.
func (s *URLShortenerService) GetLongURLFromDB(ctx context.Context, URLHash string) (models.GetURLDTO, error) {
	urlDTO, err := s.repo.GetURL(ctx, URLHash)
	return urlDTO, err
}

// GetLongURLFromReq extracts the long URL from the HTTP request body.
func (s *URLShortenerService) GetLongURLFromReq(req *http.Request) (string, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", ErrReadingBody
	}
	longURLStr := string(body)
	if longURLStr == "" {
		return "", ErrDidntGetURL
	}
	err = s.ValidateURL(longURLStr)
	if err != nil {
		return "", fmt.Errorf("got incorrect url to shorten: url=%v, err=%v: %w", longURLStr, err, ErrInvalidURL)
	}
	return longURLStr, nil
}

// GetUserID extracts the user ID from the HTTP request context.
func (s *URLShortenerService) GetUserID(req *http.Request) (string, error) {
	userIDVal := req.Context().Value(middleware.UserIDKeyName)
	if userIDVal == nil {
		return "", ErrNoUserID
	}
	userID, ok := userIDVal.(string)
	if !ok {
		return "", ErrConvertingUserID
	}
	return userID, nil
}

// GetURLHash extracts the short URL hash from the HTTP request.
func (s *URLShortenerService) GetURLHash(req *http.Request) string {
	return chi.URLParam(req, "id")
}

// GetUserURLs retrieves all URLs for a given user.
func (s *URLShortenerService) GetUserURLs(ctx context.Context, userID string) ([]models.BaseURLResponse, error) {
	if rep, ok := s.repo.(DBStorager); ok {
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

// GenerateDeleteTasks creates delete tasks for the given user and URL hashes.
func (s *URLShortenerService) GenerateDeleteTasks(userID string, urlHashes []string) {
	if _, ok := s.repo.(DBStorager); ok {
		for _, urlHash := range urlHashes {
			s.deleteCh <- models.DeleteTask{UserID: userID, URLHash: urlHash}
		}
		logger.Sugaarz.Debug("Created ", len(urlHashes), " delete tasks")
	} else {
		logger.Sugaarz.Error("not implemented error")
	}

}

// DeleteUrls runs a background process to delete URLs in batches.
func (s *URLShortenerService) DeleteUrls() {
	ticker := time.NewTicker(deleteTickerInterval)

	var (
		values       []interface{}
		placeholders []string
	)
	plInd := 1
	do := func(v []interface{}, pl []string) {
		logger.Sugaarz.Debugf("deleting urls for userID: values len=%v placeholders len=%v", len(v), len(pl))
		rowsAffected, err := s.repo.(DBStorager).DeleteURLList(v, pl)
		if err != nil {
			logger.Sugaarz.Error(err)
		} else {
			logger.Sugaarz.Debug(rowsAffected, "rows were updated on delete")
		}
	}

loop:
	for {
		select {
		case <-s.daemonsDoneCh:
			if len(values) != 0 {
				go do(values, placeholders)
			}
			break loop
		case task := <-s.deleteCh:
			values = append(values, task.UserID, task.URLHash)
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", plInd, plInd+1))
			plInd = plInd + 2

			if len(values) >= DeleteBatchSize {
				go do(values, placeholders)
				values = values[:0]
				placeholders = placeholders[:0]
				plInd = 1
			}

		case <-ticker.C:
			if len(values) != 0 {
				go do(values, placeholders)
				values = nil
				placeholders = nil
				plInd = 1
			}
		}
	}
}

// PingDB checks the connectivity to the database.
func (s *URLShortenerService) PingDB(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

// PrepareBatch prepares a batch of short URLs and validates them.
func (s *URLShortenerService) PrepareBatch(apiBatchReq []models.APIRequestPrepareBatchShURL) (
	apiBatchResp []models.APIResponsePrepareBatchShURL,
	batch []repository.URLPair,
	validationErrors []error,
	err error) {

	for _, obj := range apiBatchReq {
		validateErr := s.ValidateURL(obj.LongURL)
		if validateErr != nil {
			logger.Sugaarz.Errorw("got incorrect url to shorten in api batch: "+obj.LongURL, "err", validateErr)
			validationErrors = append(validationErrors, validateErr)
		} else {
			urlHash, errS := s.CreateShortURLHash(obj.LongURL)
			if errS != nil {
				err = errS
				return
			}
			batch = append(batch, repository.URLPair{URLHash: urlHash, LongURL: obj.LongURL})
			apiBatchResp = append(apiBatchResp, models.APIResponsePrepareBatchShURL{
				CorrelationID: obj.CorrelationID, ShortURL: s.PrepareShortURL(urlHash)})
		}
	}
	return
}

// SendDeleteTasks sends delete tasks for the given user and URL hashes to the fan in channel.
func (s *URLShortenerService) SendDeleteTasks(userID string, hashes []string) {
	s.deleteSemaphore <- struct{}{}
	defer func() {
		<-s.deleteSemaphore
	}()
	s.GenerateDeleteTasks(userID, hashes)
}
