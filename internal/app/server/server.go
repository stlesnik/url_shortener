package server

import (
	"github.com/stlesnik/url_shortener/internal/app/storage"
	"net/http"
)

type Server struct {
	router *http.ServeMux
	repo   storage.Repository
	port   string
}

func NewServer(repo storage.Repository, port string) *Server {
	s := &Server{
		router: http.NewServeMux(),
		repo:   repo,
		port:   port,
	}
	s.setupRoutes()
	return s
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.port, s.router)
}
