package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func NewConfig() *Config {
	defaultAddress := "localhost:8080"
	defaultBaseURL := "http://localhost:8080"

	serverAddrFlag := flag.String("a", "", "Address to run the server")
	baseURLFlag := flag.String("b", "", "Base URL for shortened links")

	flag.Parse()

	serverAddrEnv := os.Getenv("SERVER_ADDRESS")
	baseURLEnv := os.Getenv("BASE_URL")

	return &Config{
		ServerAddress: chooseValue(serverAddrEnv, *serverAddrFlag, defaultAddress),
		BaseURL:       chooseValue(baseURLEnv, *baseURLFlag, defaultBaseURL),
	}
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
