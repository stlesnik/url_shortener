package server

import (
	"net/http"
	_ "net/http/pprof"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/stlesnik/url_shortener/internal/app/handlers"
	"github.com/stlesnik/url_shortener/internal/app/middleware"
	"github.com/stlesnik/url_shortener/internal/app/services"
)

func (s *Server) setupRoutes() {
	service := services.New(s.repo, s.cfg)
	service.InitDeleteDaemon(s.daemonsDoneCh)
	hs := handlers.New(service)
	wrap := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.WithAuth(s.cfg,
			middleware.WithLogging(
				middleware.WithDecompress(
					middleware.WithCompress(h),
				),
			),
		)
	}
	s.router.Post("/", wrap(hs.SaveURL))
	s.router.Get("/ping", wrap(hs.PingDB))
	s.router.Get("/{id}", wrap(hs.GetLongURL))
	s.router.Post("/api/shorten", wrap(hs.APIPrepareShortURL))
	s.router.Post("/api/shorten/batch", wrap(hs.APIPrepareBatchShortURL))
	s.router.Get("/api/user/urls", wrap(hs.APIGetUserURLs))
	s.router.Delete("/api/user/urls", wrap(hs.APIDeleteUserURLs))
	s.router.Get("/api/internal/stats", middleware.WithTrustedSubnet(s.cfg, wrap(hs.APIGetStats)))

	s.router.Mount("/debug", chiMiddleware.Profiler())
}
