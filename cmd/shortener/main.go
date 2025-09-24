package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stlesnik/url_shortener/internal/app/grpc"
	"github.com/stlesnik/url_shortener/internal/app/server"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"github.com/stlesnik/url_shortener/internal/config"
	"github.com/stlesnik/url_shortener/internal/logger"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	daemonsDoneCh := make(chan struct{})
	defer close(daemonsDoneCh)
	// конфиг
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Не получилось обработать конфиг: %s", err)
		return
	}

	// логгер
	if logErr := logger.InitLogger(cfg.Environment); logErr != nil {
		panic(fmt.Errorf("logger broke: %w", logErr))
	}
	defer func() {
		if syncErr := logger.Sugaarz.Sync(); syncErr != nil {
			logger.Sugaarz.Errorw("failed to sync logger", "error", syncErr)
		}
	}()

	repo, err := services.NewRepository(cfg)
	if err != nil {
		logger.Sugaarz.Errorw("failed to create repository", "error", err)
		return
	}
	defer func() {
		if closeErr := repo.Close(); closeErr != nil {
			logger.Sugaarz.Errorw("Failed to close repository", "error", closeErr)
		}
	}()

	// Create HTTP server
	httpSrv, err := server.New(repo, cfg, daemonsDoneCh)
	if err != nil {
		logger.Sugaarz.Errorw("failed to create HTTP server", "error", err)
		return
	}

	// Create gRPC server
	grpcSrv := grpc.New(httpSrv.GetService())

	//for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Start HTTP server
	httpErrCh := make(chan error, 1)
	go func() {
		if err := httpSrv.Start(); err != nil && err != http.ErrServerClosed {
			httpErrCh <- err
		} else {
			close(httpErrCh)
		}
	}()
	log.Printf("HTTP сервер запущен на %s", cfg.ServerAddress)

	// Start gRPC server
	grpcErrCh := make(chan error, 1)
	go func() {
		if err := grpcSrv.Start(cfg.GRPCPort); err != nil {
			grpcErrCh <- err
		} else {
			close(grpcErrCh)
		}
	}()
	log.Printf("gRPC сервер запущен на порту %s", cfg.GRPCPort)

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	select {
	case sig := <-sigCh:
		logger.Sugaarz.Infow("Получен сигнал завершения", "signal", sig)
	case err := <-httpErrCh:
		if err != nil {
			logger.Sugaarz.Errorw("Ошибка HTTP сервера", "error", err)
		}
		return
	case err := <-grpcErrCh:
		if err != nil {
			logger.Sugaarz.Errorw("Ошибка gRPC сервера", "error", err)
		}
		return
	}

	logger.Sugaarz.Info("Завершение работы серверов...")

	// Shutdown HTTP server
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		logger.Sugaarz.Errorw("Ошибка при остановке HTTP сервера", "error", err)
	}

	// Shutdown gRPC server
	grpcSrv.Stop()

	logger.Sugaarz.Info("Серверы штатно остановлены")

}
