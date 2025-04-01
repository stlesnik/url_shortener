package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/stlesnik/url_shortener/internal/app/storage"
	"net/http"
)

type Server struct {
	router chi.Router
	repo   storage.Repository
	port   string
}

func NewServer(repo storage.Repository, port string) *Server {
	s := &Server{
		router: chi.NewRouter(),
		repo:   repo,
		port:   port,
	}
	s.setupRoutes()
	return s
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.port, s.router)
}
