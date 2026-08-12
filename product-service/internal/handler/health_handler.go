package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	db *pgxpool.Pool
}

func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(
		r.Context(),
		2*time.Second,
	)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		WriteError(
			w,
			http.StatusServiceUnavailable,
			"DATABASE_UNAVAILABLE",
			"Database is unavailable",
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
