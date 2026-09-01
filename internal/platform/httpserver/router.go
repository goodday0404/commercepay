package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(logger *slog.Logger, healthHandlers *HealthHandlers) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(RequestLogger(logger))
	router.Use(middleware.Recoverer)

	router.Get("/health", healthHandlers.Health)
	router.Get("/ready", healthHandlers.Ready)

	return router
}
