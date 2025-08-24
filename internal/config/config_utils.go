package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/stlesnik/url_shortener/internal/app/models"
	"os"
)

// applyFileConfig applies configuration from JSON file if specified
func applyFileConfig(cfg *Config) error {
	configPath := getConfigPath()
	if configPath == "" {
		return nil
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var fileCfg models.FileConfig
	if err := json.Unmarshal(file, &fileCfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	if fileCfg.ServerAddress != "" {
		cfg.ServerAddress = fileCfg.ServerAddress
	}
	if fileCfg.BaseURL != "" {
		cfg.BaseURL = fileCfg.BaseURL
	}
	cfg.FileStoragePath = fileCfg.FileStoragePath
	cfg.DatabaseDSN = fileCfg.DatabaseDSN
	cfg.EnableHTTPS = fileCfg.EnableHTTPS
	cfg.TrustedSubnet = fileCfg.TrustedSubnet

	return nil
}

// getConfigPath retrieves the configuration file path from env or command-line flags
func getConfigPath() string {
	if path := os.Getenv("CONFIG"); path != "" {
		return path
	}

	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	configFlag := fs.String("c", "", "Config file path")
	configFlagLong := fs.String("config", "", "Config file path")

	_ = fs.Parse(os.Args[1:])

	if *configFlag != "" {
		return *configFlag
	}
	if *configFlagLong != "" {
		return *configFlagLong
	}
	return ""
}

// parseFlags registers and parses command-line flags
func parseFlags(cfg *Config) {
	fs := flag.NewFlagSet("main", flag.ContinueOnError)

	fs.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Address to run the server")
	fs.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL for shortened links")
	fs.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "Path to file for persistent storage")
	fs.StringVar(&cfg.Environment, "e", cfg.Environment, "Environment")
	fs.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database url")
	fs.StringVar(&cfg.AuthSecretKey, "j", cfg.AuthSecretKey, "Secret key for jwt token generation")
	fs.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "Flag to enable HTTPS")
	fs.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "Flag to enable HTTPS")

	_ = fs.Parse(os.Args[1:])

}
