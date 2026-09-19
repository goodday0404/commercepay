package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/goodday0404/commercepay/internal/catalog"
	"github.com/goodday0404/commercepay/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func buildCatalog(cfg config.CatalogConfig, db *pgxpool.Pool, router chi.Router) {
	repo := catalog.NewRepository(db)
	service := catalog.NewService(repo, cfg.MaxPageSize)
	handler := catalog.NewHandler(service, cfg.DefaultPageSize)
	handler.RegisterRoutes(router)
}
