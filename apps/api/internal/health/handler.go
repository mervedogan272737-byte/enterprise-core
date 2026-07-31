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

	dbOK := h.DB.Ping(ctx) == nil
	redisOK := h.Redis.Ping(ctx).Err() == nil

	status := http.StatusOK
	if !dbOK || !redisOK {
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if dbOK && redisOK {
		w.Write([]byte(`{"status":"ok","service":"enterprise-api","database":"ok","redis":"ok"}`))
		return
	}

	w.Write([]byte(`{"status":"degraded","service":"enterprise-api","database":"check","redis":"check"}`))
}
