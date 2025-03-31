package server

import (
	"github.com/stlesnik/url_shortener/internal/app/handlers"
)

func (s *Server) setupRoutes() {
	handler := handlers.NewHandler(s.repo)
	s.router.HandleFunc("/", handler.MainHandler)
}
