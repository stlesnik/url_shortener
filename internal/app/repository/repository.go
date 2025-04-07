package repository

import "sync"

type Repository interface {
	Save(shortURL string, longURLStr string) error
	Get(shortURL string) (string, bool)
}

type InMemoryRepository struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{data: make(map[string]string)}
}

func (s *InMemoryRepository) Save(short string, long string) error {
	s.mu.Lock()
	s.data[short] = long
	s.mu.Unlock()
	return nil
}

func (s *InMemoryRepository) Get(short string) (string, bool) {
	s.mu.RLock()
	long, exists := s.data[short]
	s.mu.RUnlock()
	return long, exists
}
