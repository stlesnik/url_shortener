package server

import (
	"github.com/stlesnik/url_shortener/internal/app/handlers"
)

func (s *Server) setupRoutes() {
	hs := handlers.NewHandler(s.repo, s.cfg)
	s.router.Post("/", hs.SaveURL)
	s.router.Get("/{id}", hs.GetLongURL)
}
