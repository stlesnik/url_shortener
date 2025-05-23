package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Environment     string `env:"ENVIRONMENT"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	AuthSecretKey   string `env:"AUTH_SECRET_KEY"`
}

func New() (*Config, error) {
	cfg := &Config{}

	defaultAddress := "localhost:8080"
	defaultBaseURL := "http://localhost:8080"
	defaultFile := ""
	defaultEnvironment := "dev"
	defaultDatabaseDSN := ""
	defaultAuthSecretKey := "url_shortener_secret_key"

	flag.StringVar(&cfg.ServerAddress, "a", defaultAddress, "Address to run the server")
	flag.StringVar(&cfg.BaseURL, "b", defaultBaseURL, "Base URL for shortened links")
	flag.StringVar(&cfg.FileStoragePath, "f", defaultFile, "Path to file for persistent storage")
	flag.StringVar(&cfg.Environment, "e", defaultEnvironment, "Environment")
	flag.StringVar(&cfg.DatabaseDSN, "d", defaultDatabaseDSN, "Database url")
	flag.StringVar(&cfg.AuthSecretKey, "s", defaultAuthSecretKey, "Secret key for jwt token generation")
	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
