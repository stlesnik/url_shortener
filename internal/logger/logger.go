package logger

import (
	"strings"

	"go.uber.org/zap"
)

// Sugaarz is the global sugared logger variable used throughout the application.
var Sugaarz *zap.SugaredLogger

// InitLogger initializes the global logger variable based on the environment.
func InitLogger(env string) error {
	var logger *zap.Logger
	var err error

	switch strings.ToLower(env) {
	case "prod", "production":
		logger, err = zap.NewProduction()
	default:
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		return err
	}
	Sugaarz = logger.Sugar()
	return nil
}
