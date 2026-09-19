package httpserver

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(db *pgxpool.Pool, logger *slog.Logger) chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(RequestLogger(logger))
	router.Use(middleware.Recoverer)

	health := NewHealthHandlers(db, logger)
	health.RegisterRoutes(router)

	return router
}
