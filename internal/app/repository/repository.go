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
	defer s.mu.Unlock()
	s.data[short] = long
	return nil
}

func (s *InMemoryRepository) Get(short string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	long, exists := s.data[short]
	return long, exists
}
