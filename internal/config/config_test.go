package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Test case 1: Default values
	t.Run("Default values", func(t *testing.T) {
		resetFlags()
		os.Args = []string{"cmd"}
		cfg, err := New()
		assert.NoError(t, err)
		assert.Equal(t, "localhost:8080", cfg.ServerAddress)
		assert.Equal(t, "http://localhost:8080", cfg.BaseURL)
		assert.Equal(t, "", cfg.FileStoragePath)
		assert.Equal(t, "dev", cfg.Environment)
		assert.Equal(t, "", cfg.DatabaseDSN)
		assert.Equal(t, "url_shortener_secret_key", cfg.AuthSecretKey)
		assert.Equal(t, false, cfg.EnableHTTPS)
	})

	// Test case 2: Environment variables
	t.Run("Environment variables", func(t *testing.T) {
		resetFlags()
		os.Args = []string{"cmd"}
		assert.NoError(t, os.Setenv("SERVER_ADDRESS", "localhost:9090"))
		assert.NoError(t, os.Setenv("BASE_URL", "http://localhost:9090"))
		assert.NoError(t, os.Setenv("FILE_STORAGE_PATH", "/tmp/test.db"))
		assert.NoError(t, os.Setenv("ENVIRONMENT", "prod"))
		assert.NoError(t, os.Setenv("DATABASE_DSN", "postgres://user:password@localhost:5432/db"))
		assert.NoError(t, os.Setenv("AUTH_SECRET_KEY", "test_secret_key"))
		assert.NoError(t, os.Setenv("ENABLE_HTTPS", "true"))

		cfg, err := New()
		assert.NoError(t, err)
		assert.Equal(t, "localhost:9090", cfg.ServerAddress)
		assert.Equal(t, "http://localhost:9090", cfg.BaseURL)
		assert.Equal(t, "/tmp/test.db", cfg.FileStoragePath)
		assert.Equal(t, "prod", cfg.Environment)
		assert.Equal(t, "postgres://user:password@localhost:5432/db", cfg.DatabaseDSN)
		assert.Equal(t, "test_secret_key", cfg.AuthSecretKey)
		assert.Equal(t, true, cfg.EnableHTTPS)

		assert.NoError(t, os.Unsetenv("SERVER_ADDRESS"))
		assert.NoError(t, os.Unsetenv("BASE_URL"))
		assert.NoError(t, os.Unsetenv("FILE_STORAGE_PATH"))
		assert.NoError(t, os.Unsetenv("ENVIRONMENT"))
		assert.NoError(t, os.Unsetenv("DATABASE_DSN"))
		assert.NoError(t, os.Unsetenv("AUTH_SECRET_KEY"))
		assert.NoError(t, os.Unsetenv("ENABLE_HTTPS"))
	})

	// Test case 3: Flags
	t.Run("Flags", func(t *testing.T) {
		resetFlags()
		os.Args = []string{"cmd", "-a", "localhost:7070", "-b", "http://localhost:7070", "-f", "/tmp/test.db", "-e", "stage", "-d", "postgres://user:password@localhost:5432/testdb", "-j", "another_secret_key", "-s"}
		cfg, err := New()
		assert.NoError(t, err)
		assert.Equal(t, "localhost:7070", cfg.ServerAddress)
		assert.Equal(t, "http://localhost:7070", cfg.BaseURL)
		assert.Equal(t, "/tmp/test.db", cfg.FileStoragePath)
		assert.Equal(t, "stage", cfg.Environment)
		assert.Equal(t, "postgres://user:password@localhost:5432/testdb", cfg.DatabaseDSN)
		assert.Equal(t, "another_secret_key", cfg.AuthSecretKey)
		assert.Equal(t, true, cfg.EnableHTTPS)
	})
}

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}
