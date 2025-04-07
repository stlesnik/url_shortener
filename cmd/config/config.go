package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}

	defaultAddress := "localhost:8080"
	defaultBaseURL := "http://localhost:8080"

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	serverAddrFlag := flag.String("a", "", "Address to run the server")
	baseURLFlag := flag.String("b", "", "Base URL for shortened links")

	flag.Parse()

	return &Config{
		ServerAddress: chooseValue(cfg.ServerAddress, *serverAddrFlag, defaultAddress),
		BaseURL:       chooseValue(cfg.BaseURL, *baseURLFlag, defaultBaseURL),
	}, nil
}

func chooseValue(env, flag, def string) string {
	if env != "" {
		return env
	}
	if flag != "" {
		return flag
	}
	return def
}
