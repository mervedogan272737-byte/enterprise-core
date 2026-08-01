package health

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

func (h Handler) All(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	dbOK := h.DB != nil && h.DB.Ping(ctx) == nil
	redisOK := h.Redis != nil && h.Redis.Ping(ctx).Err() == nil

	w.Header().Set("Content-Type", "application/json")

	if !dbOK || !redisOK {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"degraded","service":"enterprise-api","database":"check","redis":"check"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok","service":"enterprise-api","database":"ok","redis":"ok"}`))
}
