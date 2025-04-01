package services

import (
	"fmt"
	"github.com/stlesnik/url_shortener/cmd/config"
)

func PrepareShortURL(urlHash string, cfg *config.Config) string {
	return fmt.Sprintf("%s/%s", cfg.BaseURL, urlHash)

}
