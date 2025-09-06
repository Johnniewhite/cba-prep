package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cbalite/backend/internal/cache"
	"github.com/cbalite/backend/internal/config"
)

func NewRateLimitMiddleware(cfg *config.RateLimitConfig, cache *cache.RedisCache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			key := fmt.Sprintf("rate_limit:%s", clientIP)
			
			ctx := r.Context()
			count, err := cache.Increment(ctx, key)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				cache.Expire(ctx, key, time.Minute)
			}

			if count > int64(cfg.RequestsPerMinute) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Rate limit exceeded"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}