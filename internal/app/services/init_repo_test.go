package services

import (
	"os"
	"testing"

	"github.com/stlesnik/url_shortener/internal/app/repository"
	"github.com/stlesnik/url_shortener/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewRepository(t *testing.T) {
	t.Run("should return InMemoryRepository when no config is provided", func(t *testing.T) {
		cfg := &config.Config{}
		repo, err := NewRepository(cfg)
		assert.NoError(t, err)
		assert.IsType(t, &repository.InMemoryRepository{}, repo)
	})

	t.Run("should return FileStorage when FileStoragePath is provided", func(t *testing.T) {
		tmpfile, err := os.CreateTemp("", "test")
		if err != nil {
			t.Fatal(err)
		}
		defer func(name string) {
			err := os.Remove(name)
			assert.NoError(t, err)
		}(tmpfile.Name())

		cfg := &config.Config{FileStoragePath: tmpfile.Name()}
		repo, err := NewRepository(cfg)
		assert.NoError(t, err)
		assert.IsType(t, &repository.FileStorage{}, repo)
		err = repo.Close()
		assert.NoError(t, err)
	})

	t.Run("should return error for invalid FileStoragePath", func(t *testing.T) {
		cfg := &config.Config{FileStoragePath: "/invalid/path/to/file"}
		_, err := NewRepository(cfg)
		assert.Error(t, err)
	})
}
