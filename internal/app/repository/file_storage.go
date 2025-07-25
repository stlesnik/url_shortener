package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/stlesnik/url_shortener/internal/app/models"
	"github.com/stlesnik/url_shortener/internal/logger"
)

// FileStorage implements a file-based URL repository.
type FileStorage struct {
	file *os.File
	data map[string]string
	mu   sync.RWMutex
}

// NewFileStorage creates a new FileStorage instance with the given file path.
func NewFileStorage(path string) (*FileStorage, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	fs := &FileStorage{
		file: file,
		data: make(map[string]string),
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var rec storedRecord
		if err := json.Unmarshal(scanner.Bytes(), &rec); err == nil {
			fs.data[rec.ShortURL] = rec.OriginalURL
		}
	}

	return fs, nil
}

// Ping checks the file storage for readiness.
func (f *FileStorage) Ping(_ context.Context) error {
	if f.data != nil {
		return nil
	}
	return fmt.Errorf("file repository is empty")
}

// SaveURL saves a URL mapping to the file storage.
func (f *FileStorage) SaveURL(ctx context.Context, short string, long string, _ string) (isDouble bool, err error) {
	select {
	case <-ctx.Done():
		logger.Sugaarz.Info("Client closed connection while in url SaveURL func")
		return false, ctx.Err()
	default:
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.data[short]; exists {
		return true, nil
	}

	f.data[short] = long

	rec := newStoredRecord(short, long)

	b, err := json.Marshal(rec)
	if err != nil {
		return false, err
	}

	_, err = f.file.Write(append(b, '\n'))
	return false, err
}

// GetURL retrieves a URL mapping from the file storage by short URL.
func (f *FileStorage) GetURL(_ context.Context, short string) (models.GetURLDTO, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	val, exists := f.data[short]
	if !exists {
		return models.GetURLDTO{}, ErrURLNotFound
	}
	return models.GetURLDTO{OriginalURL: val, IsDeleted: false}, nil
}

// Close closes the file storage.
func (f *FileStorage) Close() error {
	return f.file.Close()
}

// storedRecord represents a record stored in the file storage.
type storedRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// newStoredRecord creates a new storedRecord instance.
func newStoredRecord(shortURL, originalURL string) storedRecord {
	return storedRecord{
		UUID:        uuid.New().String(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
}
