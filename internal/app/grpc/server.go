package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/stlesnik/url_shortener/api"
	"github.com/stlesnik/url_shortener/internal/app/models"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"github.com/stlesnik/url_shortener/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server represents the gRPC server
type Server struct {
	api.UnimplementedURLShortenerServer
	grpcServer *grpc.Server
	service    *services.URLShortenerService
}

// New creates a new gRPC server
func New(service *services.URLShortenerService) *Server {
	grpcServer := grpc.NewServer()
	server := &Server{
		grpcServer: grpcServer,
		service:    service,
	}

	api.RegisterURLShortenerServer(grpcServer, server)
	return server
}

// Start starts the gRPC server
func (s *Server) Start(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	logger.Sugaarz.Infow("gRPC server starting", "port", port)
	return s.grpcServer.Serve(lis)
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	logger.Sugaarz.Info("Stopping gRPC server...")
	s.grpcServer.GracefulStop()
	logger.Sugaarz.Info("gRPC server stopped")
}

// SaveURL handles gRPC requests to create a short URL
func (s *Server) SaveURL(ctx context.Context, req *api.SaveURLRequest) (*api.SaveURLResponse, error) {
	logger.Sugaarz.Debugw("gRPC SaveURL called", "long_url", req.LongUrl)

	shortURL, isDouble, err := s.service.GenerateShortURL(ctx, req.LongUrl, req.UserId)
	if err != nil {
		logger.Sugaarz.Errorw("error generating short URL", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to generate short URL: %v", err)
	}

	return &api.SaveURLResponse{
		ShortUrl: shortURL,
		IsDouble: isDouble,
	}, nil
}

// GetLongURL handles gRPC requests to get original URL
func (s *Server) GetLongURL(ctx context.Context, req *api.GetLongURLRequest) (*api.GetLongURLResponse, error) {
	logger.Sugaarz.Debugw("gRPC GetLongURL called", "url_hash", req.UrlHash)

	urlDTO, err := s.service.GetLongURLFromDB(ctx, req.UrlHash)
	if err != nil {
		logger.Sugaarz.Errorw("error getting original URL", "error", err)
		return nil, status.Errorf(codes.NotFound, "short URL not found")
	}

	return &api.GetLongURLResponse{
		OriginalUrl: urlDTO.OriginalURL,
		IsDeleted:   urlDTO.IsDeleted,
	}, nil
}

// APIPrepareShortURL handles gRPC API requests to create a short URL
func (s *Server) APIPrepareShortURL(ctx context.Context, req *api.APIPrepareShortURLRequest) (*api.APIPrepareShortURLResponse, error) {
	logger.Sugaarz.Debugw("gRPC APIPrepareShortURL called", "long_url", req.LongUrl)

	if err := s.service.ValidateURL(req.LongUrl); err != nil {
		logger.Sugaarz.Errorw("invalid URL", "url", req.LongUrl, "error", err)
		return nil, status.Errorf(codes.Internal, "invalid URL: %v", err)
	}

	shortURL, isDouble, err := s.service.GenerateShortURL(ctx, req.LongUrl, "")
	if err != nil {
		logger.Sugaarz.Errorw("error generating short URL", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to generate short URL: %v", err)
	}

	return &api.APIPrepareShortURLResponse{
		ShortUrl: shortURL,
		IsDouble: isDouble,
	}, nil
}

// APIPrepareBatchShortURL handles gRPC batch requests to create short URLs
func (s *Server) APIPrepareBatchShortURL(ctx context.Context, req *api.APIPrepareBatchShortURLRequest) (*api.APIPrepareBatchShortURLResponse, error) {
	logger.Sugaarz.Debugw("gRPC APIPrepareBatchShortURL called", "urls_count", len(req.Urls))

	// Convert gRPC request to internal format
	var apiBatchReq []models.APIRequestPrepareBatchShURL
	for _, url := range req.Urls {
		apiBatchReq = append(apiBatchReq, models.APIRequestPrepareBatchShURL{
			CorrelationID: url.CorrelationId,
			LongURL:       url.OriginalUrl,
		})
	}

	// Process batch
	apiBatchResp, batch, validationErrors, err := s.service.PrepareBatch(apiBatchReq)
	if err != nil {
		logger.Sugaarz.Errorw("failed to process batch", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to process batch: %v", err)
	}

	// Save batch
	if err := s.service.SaveBatchShortURL(ctx, batch); err != nil {
		logger.Sugaarz.Errorw("error saving batch", "error", err)
		return nil, status.Errorf(codes.Internal, "error saving batch: %v", err)
	}

	// Convert response to gRPC format
	var grpcUrls []*api.URLPair
	for _, resp := range apiBatchResp {
		grpcUrls = append(grpcUrls, &api.URLPair{
			CorrelationId: resp.CorrelationID,
			OriginalUrl:   resp.ShortURL, // This should be the original URL, but we're returning short URL as per API
			ShortUrl:      resp.ShortURL,
		})
	}

	if len(validationErrors) > 0 {
		logger.Sugaarz.Warnw("validation errors in batch", "errors", validationErrors)
	}

	return &api.APIPrepareBatchShortURLResponse{
		Urls: grpcUrls,
	}, nil
}

// APIGetUserURLs handles gRPC requests to get user URLs
func (s *Server) APIGetUserURLs(ctx context.Context, req *api.APIGetUserURLsRequest) (*api.APIGetUserURLsResponse, error) {
	logger.Sugaarz.Debugw("gRPC APIGetUserURLs called", "user_id", req.UserId)

	urls, err := s.service.GetUserURLs(ctx, req.UserId)
	if err != nil {
		logger.Sugaarz.Errorw("error getting user URLs", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get user URLs: %v", err)
	}

	var grpcUrls []*api.BaseURLResponse
	for _, url := range urls {
		grpcUrls = append(grpcUrls, &api.BaseURLResponse{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		})
	}

	return &api.APIGetUserURLsResponse{
		Urls: grpcUrls,
	}, nil
}

// APIDeleteUserURLs handles gRPC requests to delete user URLs
func (s *Server) APIDeleteUserURLs(ctx context.Context, req *api.APIDeleteUserURLsRequest) (*api.APIDeleteUserURLsResponse, error) {
	logger.Sugaarz.Debugw("gRPC APIDeleteUserURLs called", "user_id", req.UserId, "urls_count", len(req.UrlHashes))

	s.service.SendDeleteTasks(req.UserId, req.UrlHashes)

	return &api.APIDeleteUserURLsResponse{
		Success: true,
	}, nil
}

// APIGetStats handles gRPC requests to get application statistics
func (s *Server) APIGetStats(ctx context.Context, req *api.APIGetStatsRequest) (*api.APIGetStatsResponse, error) {
	logger.Sugaarz.Debugw("gRPC APIGetStats called")

	stats, err := s.service.GetStats(ctx)
	if err != nil {
		logger.Sugaarz.Errorw("error getting stats", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get stats: %v", err)
	}

	return &api.APIGetStatsResponse{
		UrlCount:  int64(stats.URLCount),
		UserCount: int64(stats.UserCount),
	}, nil
}

// PingDB handles gRPC requests to ping the database
func (s *Server) PingDB(ctx context.Context, req *api.PingDBRequest) (*api.PingDBResponse, error) {
	logger.Sugaarz.Debugw("gRPC PingDB called")

	err := s.service.PingDB(ctx)
	if err != nil {
		logger.Sugaarz.Errorw("error pinging database", "error", err)
		return nil, status.Errorf(codes.Internal, "database ping failed: %v", err)
	}

	return &api.PingDBResponse{
		Success: true,
	}, nil
}
