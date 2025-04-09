package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/stlesnik/url_shortener/cmd/config"
	"net/http"
)

type Repository interface {
	Save(shortURL string, longURLStr string) error
	Get(shortURL string) (string, bool)
}
type Server struct {
	router chi.Router
	repo   Repository
	cfg    *config.Config
}

func NewServer(repo Repository, cfg *config.Config) *Server {
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
