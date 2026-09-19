package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/goodday0404/commercepay/internal/platform/httpserver"
)

func buildHTTPServer(httpPort string, router chi.Router) *http.Server {
	return httpserver.NewServer(
		httpserver.ServerOptions{
			Address:           ":" + httpPort,
			ReadHeaderTimeout: httpReadHeaderTimeout,
			ReadTimeout:       httpReadTimeout,
			WriteTimeout:      httpWriteTimeout,
			IdleTimeout:       httpIdleTimeout,
		},
		router,
	)
}

func runServer(server *http.Server, logger *slog.Logger) error {
	signalCtx, signalCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer signalCancel()

	serverErr := make(chan error, 1)

	go func() {
		logger.Info(
			"HTTP server starting",
			"address", server.Addr,
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

	err := server.Shutdown(shutdownCtx)
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
