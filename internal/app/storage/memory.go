package storage

type InMemoryStorage struct {
	data map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{data: make(map[string]string)}
}

func (s *InMemoryStorage) Save(short string, long string) error {
	s.data[short] = long
	return nil
}

func (s *InMemoryStorage) Get(short string) (string, bool) {
	long, exists := s.data[short]
	return long, exists
}
