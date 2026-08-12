package middleware

import (
	"net/http"
	"sync"
	"time"
)


type RateLimiter struct{
	mu sync.Mutex
	requests map[string][]time.Time
	limit int
	window time.Duration
}
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}
func (rl *RateLimiter) Middleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		ip := r.RemoteAddr

		now := time.Now()

		rl.mu.Lock()
		defer rl.mu.Unlock()

		requests := rl.requests[ip]

		cutoff := now.Add(-rl.window)

		validRequests := make(
			[]time.Time,
			0,
		)

		for _, requestTime := range requests {
			if requestTime.After(cutoff) {
				validRequests = append(
					validRequests,
					requestTime,
				)
			}
		}

		if len(validRequests) >= rl.limit {
			http.Error(
				w,
				"rate limit exceeded",
				http.StatusTooManyRequests,
			)
			return
		}

		validRequests = append(
			validRequests,
			now,
		)

		rl.requests[ip] = validRequests

		next.ServeHTTP(w, r)
	})
}