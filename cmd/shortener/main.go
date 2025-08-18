package main

import (
	"context"
	"fmt"
	"github.com/stlesnik/url_shortener/internal/app/server"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"github.com/stlesnik/url_shortener/internal/config"
	"github.com/stlesnik/url_shortener/internal/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	srv, err := server.New(repo, cfg, daemonsDoneCh)
	if err != nil {
		logger.Sugaarz.Errorw("failed to create server", "error", err)
		return
	}

	//for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	serverErrCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			serverErrCh <- err
		} else {
			close(serverErrCh)
		}
	}()
	log.Printf("Сервер запущен на %s", cfg.ServerAddress)

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	select {
	case sig := <-sigCh:
		logger.Sugaarz.Infow("Получен сигнал завершения", "signal", sig)
	case err := <-serverErrCh:
		if err != nil {
			logger.Sugaarz.Errorw("Ошибка сервера", "error", err)
		}
		return
	}

	logger.Sugaarz.Info("Завершение работы сервера...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Sugaarz.Errorw("Ошибка при остановке сервера", "error", err)
	}

	logger.Sugaarz.Info("Сервер штатно остановлен")

}
