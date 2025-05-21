package server

import (
	"github.com/stlesnik/url_shortener/internal/app/handlers"
	"github.com/stlesnik/url_shortener/internal/app/middleware"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"net/http"
)

func (s *Server) setupRoutes() {
	service := services.New(s.repo, s.cfg)
	hs := handlers.New(service)
	wrap := func(h http.HandlerFunc, createUserIfMissing bool) http.HandlerFunc {
		return middleware.WithAuth(s.cfg, createUserIfMissing,
			middleware.WithLogging(
				middleware.WithDecompress(
					middleware.WithCompress(h),
				),
			),
		)
	}
	s.router.Post("/", wrap(hs.SaveURL, true))
	s.router.Get("/ping", wrap(hs.PingDB, false))
	s.router.Get("/{id}", wrap(hs.GetLongURL, false))
	s.router.Post("/api/shorten", wrap(hs.APIPrepareShortURL, false))
	s.router.Post("/api/shorten/batch", wrap(hs.APIPrepareBatchShortURL, false))
	s.router.Get("/api/user/urls", wrap(hs.ApiGetUserURLs, false))

}
