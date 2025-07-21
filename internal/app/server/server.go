package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"github.com/stlesnik/url_shortener/internal/config"
)

type Server struct {
	router        chi.Router
	repo          services.IRepository
	cfg           *config.Config
	daemonsDoneCh chan struct{}
}

func New(repo services.IRepository, cfg *config.Config, daemonsDoneCh chan struct{}) *Server {
	s := &Server{
		router:        chi.NewRouter(),
		repo:          repo,
		cfg:           cfg,
		daemonsDoneCh: daemonsDoneCh,
	}
	s.setupRoutes()
	return s
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.cfg.ServerAddress, s.router)
}
