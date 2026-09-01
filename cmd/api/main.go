package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goodday0404/commercepay/internal/platform/config"
	"github.com/goodday0404/commercepay/internal/platform/database"
	"github.com/goodday0404/commercepay/internal/platform/httpserver"
	"github.com/goodday0404/commercepay/internal/platform/logging"
)

const (
	databaseConnectTimeout = 5 * time.Second
	shutdownTimeout        = 10 * time.Second

	httpReadHeaderTimeout = 5 * time.Second
	httpReadTimeout       = 10 * time.Second
	httpWriteTimeout      = 15 * time.Second
	httpIdleTimeout       = 60 * time.Second
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "CommercePay failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	logger, err := logging.New(logging.Options{
		Level:       cfg.LogConfig.Level,
		Format:      cfg.LogConfig.Format,
		AddSource:   cfg.LogConfig.AddSource,
		ServiceName: "commercepay-api",
		Environment: cfg.AppEnv,
	})
	if err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), databaseConnectTimeout)
	db, err := database.Open(dbCtx, cfg.DBConfig.DBUrl)
	dbCancel()

	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	defer db.Close()
	logger.Info("PostgreSQL connection established")

	healthHandlers := httpserver.NewHealthHandlers(db, logger)

	router := httpserver.NewRouter(logger, healthHandlers)
	server := httpserver.NewServer(httpserver.ServerOptions{
		Address:           ":" + cfg.HTTPPort,
		ReadHeaderTimeout: httpReadHeaderTimeout,
		ReadTimeout:       httpReadTimeout,
		WriteTimeout:      httpWriteTimeout,
		IdleTimeout:       httpIdleTimeout,
	}, router)

	signalCtx, signalCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer signalCancel()

	serverErr := make(chan error, 1)

	go func() {
		logger.Info(
			"HTTP server starting",
			"address", cfg.HTTPPort,
		)

		err := server.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}

		serverErr <- nil
	}()

	logger.Info("CommercePay API starting")

	select {
	case <-signalCtx.Done():
		logger.Info(
			"Shutdown signal received",
			"signal", signalCtx.Err().Error(),
		)
	case err := <-serverErr:
		if err != nil {
			logger.Error(
				"HTTP server stopped",
				"error", err,
			)
			return fmt.Errorf("HTTP server: %w", err)
		}

		return nil
	}

	logger.Info("graceful shutdown start")
	logger.Info("shutting down HTTP server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	if err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	err = <-serverErr
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server stopped: %w", err)
	}

	logger.Info("HTTP server stopped")

	logger.Info("graceful shutdown ends")

	return nil
}
