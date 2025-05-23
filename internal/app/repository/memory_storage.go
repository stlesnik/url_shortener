package repository

import (
	"context"
	"fmt"
	"github.com/stlesnik/url_shortener/internal/app/models"
	"sync"
)

type InMemoryRepository struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{data: make(map[string]string)}
}

func (s *InMemoryRepository) Ping(_ context.Context) error {
	if s.data != nil {
		return nil
	}
	return fmt.Errorf("in memory repository is empty")
}

func (s *InMemoryRepository) SaveURL(_ context.Context, short string, long string, _ string) (isDouble bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[short] = long
	return false, nil
}

func (s *InMemoryRepository) GetURL(_ context.Context, short string) (models.GetURLDTO, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, exists := s.data[short]
	if !exists {
		return models.GetURLDTO{}, ErrURLNotFound
	}
	return models.GetURLDTO{OriginalURL: val, IsDeleted: false}, nil
}

func (s *InMemoryRepository) Close() error { return nil }
