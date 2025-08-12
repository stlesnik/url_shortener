package repository

import (
	"context"
	"os"
	"testing"

	"github.com/stlesnik/url_shortener/internal/app/models"
	"github.com/stretchr/testify/assert"
)

func TestFileStorage(t *testing.T) {
	ctx := context.Background()

	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		err := os.Remove(name)
		assert.NoError(t, err)
	}(tmpfile.Name())

	t.Run("NewFileStorage", func(t *testing.T) {
		fs, err := NewFileStorage(tmpfile.Name())
		assert.NoError(t, err)
		assert.NotNil(t, fs)
		err = fs.Close()
		assert.NoError(t, err)
	})

	t.Run("SaveURL and GetURL", func(t *testing.T) {
		fs, _ := NewFileStorage(tmpfile.Name())
		defer func(fs *FileStorage) {
			err := fs.Close()
			assert.NoError(t, err)
		}(fs)

		shortURL := "abc"
		longURL := "http://example.com"

		_, err := fs.SaveURL(ctx, shortURL, longURL, "")
		assert.NoError(t, err)

		retrievedURL, err := fs.GetURL(ctx, shortURL)
		assert.NoError(t, err)
		assert.Equal(t, longURL, retrievedURL.OriginalURL)
	})

	t.Run("GetURL not found", func(t *testing.T) {
		fs, _ := NewFileStorage(tmpfile.Name())
		defer func(fs *FileStorage) {
			err := fs.Close()
			assert.NoError(t, err)
		}(fs)

		_, err := fs.GetURL(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Equal(t, ErrURLNotFound, err)
	})

	t.Run("Ping", func(t *testing.T) {
		fs, _ := NewFileStorage(tmpfile.Name())
		defer func(fs *FileStorage) {
			err := fs.Close()
			assert.NoError(t, err)
		}(fs)

		err := fs.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("Close", func(t *testing.T) {
		fs, _ := NewFileStorage(tmpfile.Name())
		err := fs.Close()
		assert.NoError(t, err)
	})

	t.Run("GetURL returns correct model", func(t *testing.T) {
		fs, _ := NewFileStorage(tmpfile.Name())
		defer func(fs *FileStorage) {
			err := fs.Close()
			assert.NoError(t, err)
		}(fs)

		shortURL := "def"
		longURL := "http://example.org"

		_, err := fs.SaveURL(ctx, shortURL, longURL, "")
		assert.NoError(t, err)

		expected := models.GetURLDTO{OriginalURL: longURL, IsDeleted: false}
		actual, err := fs.GetURL(ctx, shortURL)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
