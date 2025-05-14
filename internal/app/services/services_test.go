package services

import (
	"github.com/stlesnik/url_shortener/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
	storage map[string]string
	fail    bool
}

func (m *MockRepository) Save(shortURL, longURL string) error {
	if m.fail {
		return ErrSave
	}
	m.storage[shortURL] = longURL
	return nil
}

func (m *MockRepository) Get(shortURL string) (string, bool) {
	val, exists := m.storage[shortURL]
	return val, exists
}

func TestURLShortenerService_CreateSavePrepareShortURL(t *testing.T) {
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
			service := NewURLShortenerService(repo, cfg, nil)

			shortURL, errMsg := service.CreateSavePrepareShortURL(tt.longURL)

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

func TestURLShortenerService_CreateShortURLHash(t *testing.T) {
	service := NewURLShortenerService(nil, &config.Config{}, nil)

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

func TestURLShortenerService_SaveShortURL(t *testing.T) {
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
			service := NewURLShortenerService(repo, cfg, nil)

			err := service.SaveShortURL(hash, longURL)

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

func TestURLShortenerService_PrepareShortURL(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	service := NewURLShortenerService(nil, cfg, nil)
	hash := "abc123"

	result := service.PrepareShortURL(hash)
	assert.Equal(t, "http://localhost:8080/abc123", result)
}

func TestURLShortenerService_GetLongURLFromDB(t *testing.T) {
	tests := []struct {
		name        string
		hash        string
		prepopulate bool
		wantURL     string
		wantError   bool
	}{
		{"Existing URL", "abc123", true, "https://google.com", false},
		{"Non-existent URL", "badhash", false, "", true},
	}

	cfg := &config.Config{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{storage: make(map[string]string)}
			if tt.prepopulate {
				repo.storage[tt.hash] = tt.wantURL
			}
			service := NewURLShortenerService(repo, cfg, nil)

			result, err := service.GetLongURLFromDB(tt.hash)

			if tt.wantError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantURL, result)
			}
		})
	}
}
