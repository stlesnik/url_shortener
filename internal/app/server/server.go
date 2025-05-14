package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/stlesnik/url_shortener/internal/app/repository"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"github.com/stlesnik/url_shortener/internal/config"
	"net/http"
)

type Server struct {
	router chi.Router
	repo   services.Repository
	cfg    *config.Config
	db     repository.DB
}

func NewServer(repo services.Repository, cfg *config.Config, db repository.DB) *Server {
	s := &Server{
		router: chi.NewRouter(),
		repo:   repo,
		cfg:    cfg,
		db:     db,
	}
	s.setupRoutes()
	return s
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.cfg.ServerAddress, s.router)
}
