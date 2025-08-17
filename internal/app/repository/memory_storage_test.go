package repository

import (
	"context"
	"testing"

	"github.com/stlesnik/url_shortener/internal/app/models"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("NewInMemoryRepository", func(t *testing.T) {
		repo := NewInMemoryRepository()
		assert.NotNil(t, repo)
		assert.NotNil(t, repo.data)
	})

	t.Run("Ping", func(t *testing.T) {
		repo := NewInMemoryRepository()
		err := repo.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("SaveURL and GetURL", func(t *testing.T) {
		repo := NewInMemoryRepository()
		shortURL := "abc"
		longURL := "http://example.com"

		_, err := repo.SaveURL(ctx, shortURL, longURL, "")
		assert.NoError(t, err)

		retrievedURL, err := repo.GetURL(ctx, shortURL)
		assert.NoError(t, err)
		assert.Equal(t, longURL, retrievedURL.OriginalURL)
	})

	t.Run("GetURL not found", func(t *testing.T) {
		repo := NewInMemoryRepository()
		_, err := repo.GetURL(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Equal(t, ErrURLNotFound, err)
	})

	t.Run("Close", func(t *testing.T) {
		repo := NewInMemoryRepository()
		err := repo.Close()
		assert.NoError(t, err)
	})

	t.Run("Ping after Close", func(t *testing.T) {
		repo := NewInMemoryRepository()
		err := repo.Close()
		assert.NoError(t, err)
		err = repo.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("SaveURL after Close", func(t *testing.T) {
		repo := NewInMemoryRepository()
		err := repo.Close()
		assert.NoError(t, err)
		_, err = repo.SaveURL(ctx, "a", "b", "c")
		assert.NoError(t, err)
	})

	t.Run("GetURL after Close", func(t *testing.T) {
		repo := NewInMemoryRepository()
		err := repo.Close()
		assert.NoError(t, err)
		_, err = repo.GetURL(ctx, "a")
		assert.Error(t, err)
	})

	t.Run("GetURL returns correct model", func(t *testing.T) {
		repo := NewInMemoryRepository()
		shortURL := "abc"
		longURL := "http://example.com"

		_, err := repo.SaveURL(ctx, shortURL, longURL, "")
		assert.NoError(t, err)

		expected := models.GetURLDTO{OriginalURL: longURL, IsDeleted: false}
		actual, err := repo.GetURL(ctx, shortURL)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
