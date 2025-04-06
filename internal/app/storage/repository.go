package storage

type Repository interface {
	Save(shortURL string, longURLStr string) error
	Get(shortURL string) (string, bool)
}

type InMemoryRepository struct {
	data map[string]string
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{data: make(map[string]string)}
}

func (s *InMemoryRepository) Save(short string, long string) error {
	s.data[short] = long
	return nil
}

func (s *InMemoryRepository) Get(short string) (string, bool) {
	long, exists := s.data[short]
	return long, exists
}
