package server

import (
	"context"
	"errors"
	"net/http"

	"golang.org/x/crypto/acme/autocert"

	"github.com/go-chi/chi/v5"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"github.com/stlesnik/url_shortener/internal/config"
)

// Server represents the HTTP server for the URL shortener service.
type Server struct {
	httpServer    *http.Server
	router        chi.Router
	repo          services.Storager
	cfg           *config.Config
	daemonsDoneCh chan struct{}
	service       *services.URLShortenerService
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

	// Initialize service
	s.service = services.New(repo, cfg)
	s.service.InitDeleteDaemon(daemonsDoneCh)

	s.setupRoutes()

	s.httpServer = &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: s.router,
	}

	if cfg.EnableHTTPS {
		manager := &autocert.Manager{
			Cache:      autocert.DirCache("cache-dir"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(),
		}
		s.httpServer.TLSConfig = manager.TLSConfig()
	}

	return s, nil
}

// Start runs the HTTP server.
func (s *Server) Start() error {
	if s.cfg.EnableHTTPS {
		return s.httpServer.ListenAndServeTLS("", "")
	}
	return s.httpServer.ListenAndServe()
}

// Shutdown stops the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// GetService returns the URLShortenerService instance
func (s *Server) GetService() *services.URLShortenerService {
	return s.service
}
