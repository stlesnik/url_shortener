package services

type Repository interface {
	Ping() error
	Save(shortURL string, longURLStr string) error
	Get(shortURL string) (string, bool)
	Close() error
}
