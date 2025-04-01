package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/stlesnik/url_shortener/cmd/config"
	"github.com/stlesnik/url_shortener/internal/app/storage"
	"net/http"
)

type Server struct {
	router chi.Router
	repo   storage.Repository
	cfg    *config.Config
}

func NewServer(repo storage.Repository, cfg *config.Config) *Server {
	s := &Server{
		router: chi.NewRouter(),
		repo:   repo,
		cfg:    cfg,
	}
	s.setupRoutes()
	return s
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.cfg.ServerAddress, s.router)
}
