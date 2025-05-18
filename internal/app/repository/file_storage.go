package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"os"
	"sync"
)

type FileStorage struct {
	file *os.File
	data map[string]string
	mu   sync.RWMutex
}

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

func (f *FileStorage) Ping(_ context.Context) error {
	if f.data != nil {
		return nil
	}
	return fmt.Errorf("file repository is empty")
}

func (f *FileStorage) Save(ctx context.Context, short string, long string) (isDouble bool, err error) {
	select {
	case <-ctx.Done():
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

func (f *FileStorage) Get(_ context.Context, short string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	val, exists := f.data[short]
	if !exists {
		return "", ErrURLNotFound
	}
	return val, nil
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}

type storedRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func newStoredRecord(shortURL, originalURL string) storedRecord {
	return storedRecord{
		UUID:        uuid.New().String(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
}
