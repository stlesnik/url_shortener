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
}

func NewConfig() (*Config, error) {
	cfg := &Config{}

	defaultAddress := "localhost:8080"
	defaultBaseURL := "http://localhost:8080"
	defaultFile := "storage.json"
	defaultEnvironment := "dev"

	flag.StringVar(&cfg.ServerAddress, "a", defaultAddress, "Address to run the server")
	flag.StringVar(&cfg.BaseURL, "b", defaultBaseURL, "Base URL for shortened links")
	flag.StringVar(&cfg.FileStoragePath, "f", defaultFile, "Path to file for persistent storage")
	flag.StringVar(&cfg.Environment, "e", defaultEnvironment, "Environment")
	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
