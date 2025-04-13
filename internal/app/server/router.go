package server

import (
	"github.com/stlesnik/url_shortener/internal/app/handlers"
	"github.com/stlesnik/url_shortener/internal/app/middleware"
	"github.com/stlesnik/url_shortener/internal/app/services"
)

func (s *Server) setupRoutes() {
	service := services.NewURLShortenerService(s.repo, s.cfg)
	hs := handlers.NewHandler(service)
	s.router.Post("/", middleware.WithLogging(hs.SaveURL))
	s.router.Get("/{id}", middleware.WithLogging(hs.GetLongURL))
}
