package services

import (
	"context"
	"github.com/stlesnik/url_shortener/internal/app/models"
	"github.com/stlesnik/url_shortener/internal/app/repository"
	"github.com/stlesnik/url_shortener/internal/config"
	"github.com/stlesnik/url_shortener/internal/logger"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
	storage map[string]string
	fail    bool
}

func (m *MockRepository) Ping(_ context.Context) error { return nil }

func (m *MockRepository) SaveURL(_ context.Context, shortURL, longURL string, _ string) (bool, error) {
	if m.fail {
		return false, ErrServiceSave
	}
	m.storage[shortURL] = longURL
	return false, nil
}

func (m *MockRepository) GetURL(_ context.Context, shortURL string) (models.GetURLDTO, error) {
	val, exists := m.storage[shortURL]
	if !exists {
		return models.GetURLDTO{}, repository.ErrURLNotFound
	}
	return models.GetURLDTO{OriginalURL: val, IsDeleted: false}, nil
}

func (m *MockRepository) GetURLList(_ context.Context, _ string) ([]models.BaseURLDTO, error) {
	return nil, nil
}

func (m *MockRepository) Close() error {
	return nil
}

func TestServices_CreateSavePrepareShortURL(t *testing.T) {
	tests := []struct {
		name        string
		longURL     string
		wantError   bool
		repoFailure bool
	}{
		{
			name:      "Valid URL",
			longURL:   "https://google.com",
			wantError: false,
		},
		{
			name:        "Repository failure",
			longURL:     "https://google.com",
			wantError:   true,
			repoFailure: true,
		},
	}

	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{
				storage: make(map[string]string),
				fail:    tt.repoFailure,
			}
			service := New(repo, cfg, nil)

			shortURL, _, errMsg := service.CreateSavePrepareShortURL(context.Background(), tt.longURL, "")

			if tt.wantError {
				assert.NotEmpty(t, errMsg)
				assert.Empty(t, shortURL)
			} else {
				assert.Empty(t, errMsg)
				assert.Contains(t, shortURL, cfg.BaseURL)
				assert.NotEmpty(t, shortURL)
			}
		})
	}
}

func TestServices_CreateShortURLHash(t *testing.T) {
	service := New(nil, &config.Config{}, nil)

	t.Run("Hash generation", func(t *testing.T) {
		url1 := "https://google.com"
		url2 := "https://google.com/"

		hash1, err1 := service.CreateShortURLHash(url1)
		hash2, err2 := service.CreateShortURLHash(url2)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2)
		assert.NotEmpty(t, hash1)
		assert.NotEmpty(t, hash2)
	})
}

func TestServices_SaveShortURL(t *testing.T) {
	tests := []struct {
		name        string
		repoFailure bool
		wantError   bool
	}{
		{"Successful save", false, false},
		{"Repository failure", true, true},
	}

	cfg := &config.Config{}
	longURL := "https://google.com"
	hash := "abc123"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{
				storage: make(map[string]string),
				fail:    tt.repoFailure,
			}
			service := New(repo, cfg, nil)

			_, err := service.SaveShortURL(context.Background(), hash, longURL, "")

			if tt.wantError {
				assert.Error(t, err)
				assert.Empty(t, repo.storage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, longURL, repo.storage[hash])
			}
		})
	}
}

func TestServices_SaveBatchShortURL(t *testing.T) {
	tests := []struct {
		name        string
		repoFailure bool
		wantError   bool
	}{
		{"Successful save", false, false},
		{"Successful save to db", false, false},
		{"Repository failure", true, true},
	}

	cfg := &config.Config{}
	err := logger.InitLogger(cfg.Environment)
	require.NoError(t, err)
	urlPairList := []repository.URLPair{
		{URLHash: "https://google.com", LongURL: "abc123"},
		{URLHash: "https://google1.com", LongURL: "abc1234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{
				storage: make(map[string]string),
				fail:    tt.repoFailure,
			}

			service := New(repo, cfg, nil)

			err := service.SaveBatchShortURL(context.Background(), urlPairList)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServices_PrepareShortURL(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	service := New(nil, cfg, nil)
	hash := "abc123"

	result := service.PrepareShortURL(hash)
	assert.Equal(t, "http://localhost:8080/abc123", result)
}

func TestServices_GetLongURLFromDB(t *testing.T) {
	tests := []struct {
		name        string
		hash        string
		prepopulate bool
		wantRes     models.GetURLDTO
		wantError   bool
	}{
		{"Existing URL", "abc123", true, models.GetURLDTO{OriginalURL: "https://google.com", IsDeleted: false}, false},
		{"Non-existent URL", "badhash", false, models.GetURLDTO{}, true},
	}

	cfg := &config.Config{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{storage: make(map[string]string)}
			if tt.prepopulate {
				repo.storage[tt.hash] = tt.wantRes.OriginalURL
			}
			service := New(repo, cfg, nil)

			result, err := service.GetLongURLFromDB(context.Background(), tt.hash)

			if tt.wantError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRes, result)
			}
		})
	}
}
