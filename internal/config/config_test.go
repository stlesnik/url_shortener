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
		assert.Equal(t, "", cfg.TrustedSubnet)
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
		assert.NoError(t, os.Setenv("TRUSTED_SUBNET", "128.0.0.0"))

		cfg, err := New()
		assert.NoError(t, err)
		assert.Equal(t, "localhost:9090", cfg.ServerAddress)
		assert.Equal(t, "http://localhost:9090", cfg.BaseURL)
		assert.Equal(t, "/tmp/test.db", cfg.FileStoragePath)
		assert.Equal(t, "prod", cfg.Environment)
		assert.Equal(t, "postgres://user:password@localhost:5432/db", cfg.DatabaseDSN)
		assert.Equal(t, "test_secret_key", cfg.AuthSecretKey)
		assert.Equal(t, true, cfg.EnableHTTPS)
		assert.Equal(t, "128.0.0.0", cfg.TrustedSubnet)

		assert.NoError(t, os.Unsetenv("SERVER_ADDRESS"))
		assert.NoError(t, os.Unsetenv("BASE_URL"))
		assert.NoError(t, os.Unsetenv("FILE_STORAGE_PATH"))
		assert.NoError(t, os.Unsetenv("ENVIRONMENT"))
		assert.NoError(t, os.Unsetenv("DATABASE_DSN"))
		assert.NoError(t, os.Unsetenv("AUTH_SECRET_KEY"))
		assert.NoError(t, os.Unsetenv("ENABLE_HTTPS"))
		assert.NoError(t, os.Unsetenv("TRUSTED_SUBNET"))
	})

	// Test case 3: Flags
	t.Run("Flags", func(t *testing.T) {
		resetFlags()
		os.Args = []string{"cmd", "-a", "localhost:7070", "-b", "http://localhost:7070", "-f", "/tmp/test.db", "-e", "stage", "-d", "postgres://user:password@localhost:5432/testdb", "-j", "another_secret_key", "-t", "10.0.0.0", "-s"}
		cfg, err := New()
		assert.NoError(t, err)
		assert.Equal(t, "localhost:7070", cfg.ServerAddress)
		assert.Equal(t, "http://localhost:7070", cfg.BaseURL)
		assert.Equal(t, "/tmp/test.db", cfg.FileStoragePath)
		assert.Equal(t, "stage", cfg.Environment)
		assert.Equal(t, "postgres://user:password@localhost:5432/testdb", cfg.DatabaseDSN)
		assert.Equal(t, "another_secret_key", cfg.AuthSecretKey)
		assert.Equal(t, true, cfg.EnableHTTPS)
		assert.Equal(t, "10.0.0.0", cfg.TrustedSubnet)
	})

	// Test case 4: Config file
	t.Run("Config file", func(t *testing.T) {
		resetFlags()
		os.Args = []string{"cmd"}

		// Create temporary config file
		configContent := `{
			"server_address": "127.0.0.1:8080",
			"base_url": "http://config.test",
			"file_storage_path": "/tmp/config_db.json",
			"database_dsn": "postgres://config:pass@localhost:5432/db",
			"enable_https": true,
			"trusted_subnet": "128.0.0.0"
		}`

		tmpFile, err := os.CreateTemp("", "config*.json")
		assert.NoError(t, err)
		defer func(name string) {
			err := os.Remove(name)
			assert.NoError(t, err)
		}(tmpFile.Name())

		_, err = tmpFile.Write([]byte(configContent))
		assert.NoError(t, err)
		err = tmpFile.Close()
		assert.NoError(t, err)

		t.Setenv("CONFIG", tmpFile.Name())

		cfg, err := New()
		assert.NoError(t, err)
		assert.Equal(t, "127.0.0.1:8080", cfg.ServerAddress)
		assert.Equal(t, "http://config.test", cfg.BaseURL)
		assert.Equal(t, "/tmp/config_db.json", cfg.FileStoragePath)
		assert.Equal(t, "postgres://config:pass@localhost:5432/db", cfg.DatabaseDSN)
		assert.True(t, cfg.EnableHTTPS)
		assert.Equal(t, "128.0.0.0", cfg.TrustedSubnet)

		// Fields not in config file should keep defaults
		assert.Equal(t, "dev", cfg.Environment)
		assert.Equal(t, "url_shortener_secret_key", cfg.AuthSecretKey)
	})

	// Test case 5: Empty values in config file
	t.Run("Config file with empty values", func(t *testing.T) {
		resetFlags()
		os.Args = []string{"cmd"}

		configContent := `{
			"server_address": "",
			"base_url": "",
			"file_storage_path": "",
			"database_dsn": "",
			"enable_https": false
		}`

		tmpFile, err := os.CreateTemp("", "config*.json")
		assert.NoError(t, err)
		defer func(name string) {
			err := os.Remove(name)
			assert.NoError(t, err)
		}(tmpFile.Name())

		_, err = tmpFile.Write([]byte(configContent))
		assert.NoError(t, err)
		err = tmpFile.Close()
		assert.NoError(t, err)

		t.Setenv("CONFIG", tmpFile.Name())

		cfg, err := New()
		assert.NoError(t, err)

		// Empty ServerAddress/BaseURL should be ignored
		assert.Equal(t, "localhost:8080", cfg.ServerAddress)
		assert.Equal(t, "http://localhost:8080", cfg.BaseURL)

		// Empty values should be applied for other fields
		assert.Equal(t, "", cfg.FileStoragePath)
		assert.Equal(t, "", cfg.DatabaseDSN)
		assert.False(t, cfg.EnableHTTPS)
	})

	// Test case 6: Priority - flags > env > config file
	t.Run("Priority order", func(t *testing.T) {
		resetFlags()

		// Create config file
		configContent := `{
			"server_address": "127.0.0.1:8080",
			"base_url": "http://config.test",
			"file_storage_path": "/tmp/config_db.json",
			"database_dsn": "postgres://config:pass@localhost:5432/db",
			"enable_https": true
		}`

		tmpFile, err := os.CreateTemp("", "config*.json")
		assert.NoError(t, err)
		defer func(name string) {
			err := os.Remove(name)
			assert.NoError(t, err)
		}(tmpFile.Name())

		_, err = tmpFile.Write([]byte(configContent))
		assert.NoError(t, err)
		err = tmpFile.Close()
		assert.NoError(t, err)

		// Set environment variables (should override config file)
		t.Setenv("CONFIG", tmpFile.Name())
		t.Setenv("SERVER_ADDRESS", "env.host:9090")
		t.Setenv("FILE_STORAGE_PATH", "/env/path.db")
		t.Setenv("ENABLE_HTTPS", "false")

		// Set command-line flags (should override everything)
		os.Args = []string{"cmd",
			"-a", "flag.host:7070",
			"-f", "/flag/path.db",
			"-s",
		}

		cfg, err := New()
		assert.NoError(t, err)

		// Flags should have highest priority
		assert.Equal(t, "flag.host:7070", cfg.ServerAddress)
		assert.Equal(t, "/flag/path.db", cfg.FileStoragePath)
		assert.True(t, cfg.EnableHTTPS) // -s flag sets to true

		// Environment should override config file
		// (BASE_URL not set in env or flags -> should come from config)
		assert.Equal(t, "http://config.test", cfg.BaseURL)

		// DatabaseDSN not set in env or flags -> from config
		assert.Equal(t, "postgres://config:pass@localhost:5432/db", cfg.DatabaseDSN)

		// Fields not in config should keep defaults
		assert.Equal(t, "dev", cfg.Environment)
	})

	// Test case 7: Config file not found
	t.Run("Config file not found", func(t *testing.T) {
		resetFlags()
		os.Args = []string{"cmd"}
		t.Setenv("CONFIG", "/non/existing/path.json")

		_, err := New()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read config file")
	})

	// Test case 8: Invalid config file
	t.Run("Invalid config file", func(t *testing.T) {
		resetFlags()
		os.Args = []string{"cmd"}

		tmpFile, err := os.CreateTemp("", "config*.json")
		assert.NoError(t, err)
		defer func(name string) {
			err := os.Remove(name)
			assert.NoError(t, err)
		}(tmpFile.Name())

		_, err = tmpFile.Write([]byte("{invalid json}"))
		assert.NoError(t, err)
		err = tmpFile.Close()
		assert.NoError(t, err)

		t.Setenv("CONFIG", tmpFile.Name())

		_, err = New()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse config file")
	})
}

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}
