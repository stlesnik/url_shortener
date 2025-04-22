package server

import (
	"github.com/stlesnik/url_shortener/internal/app/handlers"
	"github.com/stlesnik/url_shortener/internal/app/middleware"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"net/http"
)

func (s *Server) setupRoutes() {
	service := services.NewURLShortenerService(s.repo, s.cfg)
	hs := handlers.NewHandler(service)
	wrap := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.WithLogging(
			middleware.WithDecompress(
				middleware.WithCompress(h),
			),
		)
	}
	s.router.Post("/", wrap(hs.SaveURL))
	s.router.Get("/{id}", wrap(hs.GetLongURL))
	s.router.Post("/api/shorten", wrap(hs.APIPrepareShortURL))
}
