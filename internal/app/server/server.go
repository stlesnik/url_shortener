package server

import (
	"errors"
	"golang.org/x/crypto/acme/autocert"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"github.com/stlesnik/url_shortener/internal/config"
)

// Server represents the HTTP server for the URL shortener service.
type Server struct {
	router        chi.Router
	repo          services.Storager
	cfg           *config.Config
	daemonsDoneCh chan struct{}
}

// New creates a new Server instance with the given repository, config, and daemons channel.
func New(repo services.Storager, cfg *config.Config, daemonsDoneCh chan struct{}) (*Server, error) {
	if repo == nil || cfg == nil || daemonsDoneCh == nil {
		return nil, errors.New("repository, config, or daemons channel is nil")
	}

	s := &Server{
		router:        chi.NewRouter(),
		repo:          repo,
		cfg:           cfg,
		daemonsDoneCh: daemonsDoneCh,
	}
	s.setupRoutes()
	return s, nil
}

// Start runs the HTTP server.
func (s *Server) Start() error {
	server := &http.Server{
		Addr:    s.cfg.ServerAddress,
		Handler: s.router,
	}
	if s.cfg.EnableHTTPS {
		manager := &autocert.Manager{
			Cache:      autocert.DirCache("cache-dir"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(),
		}
		server.TLSConfig = manager.TLSConfig()
		return server.ListenAndServeTLS("", "")
	}
	return server.ListenAndServe()
}
