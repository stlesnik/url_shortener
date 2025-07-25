package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/stlesnik/url_shortener/internal/app/models"
)

// InMemoryRepository implements an in-memory URL repository.
type InMemoryRepository struct {
	data map[string]string
	mu   sync.RWMutex
}

// NewInMemoryRepository creates a new in-memory repository.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{data: make(map[string]string)}
}

// Ping checks the in-memory repository for readiness.
func (s *InMemoryRepository) Ping(_ context.Context) error {
	if s.data != nil {
		return nil
	}
	return fmt.Errorf("in memory repository is empty")
}

// SaveURL saves a URL mapping to the in-memory repository.
func (s *InMemoryRepository) SaveURL(_ context.Context, short string, long string, _ string) (isDouble bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[short] = long
	return false, nil
}

// GetURL retrieves a URL mapping from the in-memory repository by short URL.
func (s *InMemoryRepository) GetURL(_ context.Context, short string) (models.GetURLDTO, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, exists := s.data[short]
	if !exists {
		return models.GetURLDTO{}, ErrURLNotFound
	}
	return models.GetURLDTO{OriginalURL: val, IsDeleted: false}, nil
}

// Close closes the in-memory repository.
func (s *InMemoryRepository) Close() error { return nil }
