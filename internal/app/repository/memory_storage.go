package repository

import (
	"context"
	"fmt"
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

func (s *InMemoryRepository) Save(_ context.Context, short string, long string) (isDouble bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[short] = long
	return false, nil
}

func (s *InMemoryRepository) Get(_ context.Context, short string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, exists := s.data[short]
	if !exists {
		return "", ErrURLNotFound
	}
	return val, nil
}

func (s *InMemoryRepository) Close() error { return nil }
