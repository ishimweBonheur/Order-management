package middleware

import (
"context"
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
ip := r.RemoteAddr

key := rl.keyPrefix + ip

ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
defer cancel()

windowSeconds := int(rl.window.Seconds())

pipe := rl.client.Pipeline()

incr := pipe.Incr(ctx, key)
pipe.Expire(ctx, key, rl.window)

if _, err := pipe.Exec(ctx); err != nil {
http.Error(
w,
"rate limiter unavailable",
http.StatusServiceUnavailable,
)
return
}

count := incr.Val()

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
