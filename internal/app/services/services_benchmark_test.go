package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stlesnik/url_shortener/internal/app/models"
	"github.com/stlesnik/url_shortener/internal/app/repository"
	"github.com/stlesnik/url_shortener/internal/config"
)

type BenchmarkMockRepository struct {
	storage map[string]string
}

func (m *BenchmarkMockRepository) Ping(_ context.Context) error { return nil }
func (m *BenchmarkMockRepository) SaveURL(_ context.Context, shortURL, longURL string, _ string) (bool, error) {
	m.storage[shortURL] = longURL
	return false, nil
}
func (m *BenchmarkMockRepository) GetURL(_ context.Context, shortURL string) (models.GetURLDTO, error) {
	val, exists := m.storage[shortURL]
	if !exists {
		return models.GetURLDTO{}, repository.ErrURLNotFound
	}
	return models.GetURLDTO{OriginalURL: val, IsDeleted: false}, nil
}
func (m *BenchmarkMockRepository) Close() error { return nil }

func BenchmarkCreateShortURLHash(b *testing.B) {
	service := New(nil, &config.Config{})
	url := "https://example.com/benchmark"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.CreateShortURLHash(url)
	}
}

func BenchmarkSaveShortURL(b *testing.B) {
	repo := &BenchmarkMockRepository{storage: make(map[string]string)}
	service := New(repo, &config.Config{})
	url := "https://example.com/benchmark"
	hash, _ := service.CreateShortURLHash(url)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.SaveShortURL(context.Background(), hash, url, "")
	}
}

func BenchmarkGenerateShortURL(b *testing.B) {
	repo := &BenchmarkMockRepository{storage: make(map[string]string)}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	service := New(repo, cfg)
	url := "https://example.com/benchmark"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = service.GenerateShortURL(context.Background(), url, "")
	}
}

func BenchmarkPrepareShortURL(b *testing.B) {
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	service := New(nil, cfg)
	hash := "benchhash"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.PrepareShortURL(hash)
	}
}

func BenchmarkGetLongURLFromDB(b *testing.B) {
	repo := &BenchmarkMockRepository{storage: make(map[string]string)}
	service := New(repo, &config.Config{})
	url := "https://example.com/benchmark"
	hash, _ := service.CreateShortURLHash(url)
	_, err := repo.SaveURL(context.Background(), hash, url, "")
	if err != nil {
		require.NoError(b, err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetLongURLFromDB(context.Background(), hash)
	}
}
