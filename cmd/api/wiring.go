package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/goodday0404/commercepay/internal/catalog"
	"github.com/goodday0404/commercepay/internal/platform/httpserver"
)

type handlers struct {
	Health  *httpserver.HealthHandlers
	Catalog *catalog.Handler
}

func buildHandlers(deps *dependencies) handlers {
	healthHandler := httpserver.NewHealthHandlers(deps.DB, deps.Logger)

	catalogRepo := catalog.NewRepository(deps.DB)
	catalogService := catalog.NewService(catalogRepo)
	catalogHandler := catalog.NewHandler(catalogService)

	return handlers{
		Health:  healthHandler,
		Catalog: catalogHandler,
	}
}

func registerRoutes(r chi.Router, h handlers) {
	h.Health.RegisterRoutes(r)
	h.Catalog.RegisterRoutes(r)
}
