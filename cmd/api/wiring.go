package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/goodday0404/commercepay/internal/catalog"
	"github.com/jackc/pgx/v5/pgxpool"
)

func buildCatalog(db *pgxpool.Pool, router chi.Router) {
	repo := catalog.NewRepository(db)
	service := catalog.NewService(repo)
	handler := catalog.NewHandler(service)
	handler.RegisterRoutes(router)
}
