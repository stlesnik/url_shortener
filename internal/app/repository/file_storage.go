package repository

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type FileStorage struct {
	file  *os.File
	data  map[string]string
	mu    sync.RWMutex
	index int
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
			fs.index++
		}
	}

	return fs, nil
}
func (f *FileStorage) Save(short string, long string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.data[short]; exists {
		return nil
	}

	f.data[short] = long
	f.index++

	rec := storedRecord{
		UUID:        fmt.Sprintf("%d", f.index),
		ShortURL:    short,
		OriginalURL: long,
	}

	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	_, err = f.file.Write(append(b, '\n'))
	return err
}

func (f *FileStorage) Get(short string) (string, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	val, ok := f.data[short]
	return val, ok
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}

type storedRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
