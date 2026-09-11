package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/goodday0404/commercepay/internal/platform/config"
	"github.com/goodday0404/commercepay/internal/platform/database"
	"github.com/goodday0404/commercepay/internal/platform/httpserver"
	"github.com/goodday0404/commercepay/internal/platform/logging"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	databaseConnectTimeout = 5 * time.Second
	shutdownTimeout        = 10 * time.Second

	httpReadHeaderTimeout = 5 * time.Second
	httpReadTimeout       = 10 * time.Second
	httpWriteTimeout      = 15 * time.Second
	httpIdleTimeout       = 60 * time.Second
)

type dependencies struct {
	Logger *slog.Logger
	DB     *pgxpool.Pool
	Router chi.Router
}

func buildDependencies(cfg config.Config) (*dependencies, error) {
	logger, err := logging.New(logging.Options{
		Level:       cfg.LogConfig.Level,
		Format:      cfg.LogConfig.Format,
		AddSource:   cfg.LogConfig.AddSource,
		ServiceName: "commercepay-api",
		Environment: cfg.AppEnv,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize logger: %w", err)
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), databaseConnectTimeout)
	db, err := database.Open(dbCtx, cfg.DBConfig.DBUrl)
	dbCancel()

	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	logger.Info("PostgreSQL connection established")

	router := httpserver.NewRouter(db, logger)

	return &dependencies{
		Logger: logger,
		DB:     db,
		Router: router,
	}, nil
}

func (d *dependencies) Close() {
	d.DB.Close()
}
