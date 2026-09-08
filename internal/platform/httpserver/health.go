package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandlers struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewHealthHandlers(db *pgxpool.Pool, logger *slog.Logger) *HealthHandlers {
	return &HealthHandlers{
		db:     db,
		logger: logger,
	}
}

func (h *HealthHandlers) Health(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HealthHandlers) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		h.logger.Warn(
			"readiness check failed",
			"error", err,
		)

		WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
