package server

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// visitor holds the rate limiter and last-seen time for a single IP.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// rateLimiter tracks per-IP visitors and evicts stale entries periodically.
type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	// r is the number of tokens added per second; b is the burst size.
	r rate.Limit
	b int
}

func newRateLimiter(r rate.Limit, b int) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		r:        r,
		b:        b,
	}
	go rl.cleanupLoop()
	return rl
}

// getLimiter returns (or creates) the rate limiter for the given IP.
func (rl *rateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	v, ok := rl.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(rl.r, rl.b)}
		rl.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupLoop removes visitors that haven't been seen in the last 5 minutes.
func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 5*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware returns an http.Handler middleware that enforces the rate limit.
func (rl *rateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realIP(r)
		if !rl.getLimiter(ip).Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware is a ready-to-use middleware: 10 req/s, burst of 20.
var RateLimitMiddleware = newRateLimiter(10, 20).Middleware
