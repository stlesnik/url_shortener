package services

type Repository interface {
	Save(shortURL string, longURLStr string) error
	Get(shortURL string) (string, bool)
}
