package repository

import (
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

func (s *InMemoryRepository) Ping() error {
	if s.data != nil {
		return nil
	}
	return fmt.Errorf("in memory repository is empty")
}

func (s *InMemoryRepository) Save(short string, long string) (isDouble bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[short] = long
	return false, nil
}

func (s *InMemoryRepository) Get(short string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	long, exists := s.data[short]
	return long, exists
}

func (s *InMemoryRepository) Close() error { return nil }
