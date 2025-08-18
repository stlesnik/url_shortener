package config

import (
	"github.com/caarlos0/env/v6"
)

// Config holds application configuration.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Environment     string `env:"ENVIRONMENT"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	AuthSecretKey   string `env:"AUTH_SECRET_KEY"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS"`
}

// New creates a new Config by parsing configuration file, environment variables, and flags.
func New() (*Config, error) {
	cfg := &Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
		Environment:     "dev",
		DatabaseDSN:     "",
		AuthSecretKey:   "url_shortener_secret_key",
		EnableHTTPS:     false,
	}

	if err := applyFileConfig(cfg); err != nil {
		return nil, err
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	parseFlags(cfg)

	return cfg, nil
}
