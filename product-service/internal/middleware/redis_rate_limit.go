package middleware

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	client    *redis.Client
	limit     int
	window    time.Duration
	keyPrefix string
}

func NewRedisRateLimiter(
	client *redis.Client,
	limit int,
	window time.Duration,
) *RedisRateLimiter {
	return &RedisRateLimiter{
		client:    client,
		limit:     limit,
		window:    window,
		keyPrefix: "rate_limit:",
	}
}

func (rl *RedisRateLimiter) Middleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		key := rl.keyPrefix + ip

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		windowSeconds := int(rl.window.Seconds())

		count, err := rl.client.Incr(ctx, key).Result()
		if err == nil && count == 1 {
			err = rl.client.Expire(ctx, key, rl.window).Err()
		}

		if err != nil {
			http.Error(
				w,
				"rate limiter unavailable",
				http.StatusServiceUnavailable,
			)
			return
		}

		if int(count) > rl.limit {
			w.Header().Set("Retry-After", strconv.Itoa(windowSeconds))
			http.Error(
				w,
				"rate limit exceeded",
				http.StatusTooManyRequests,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}
