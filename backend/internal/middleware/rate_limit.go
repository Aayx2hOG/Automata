package middleware

import (
	"net/http"
	"net/netip"
	"os"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mu          sync.Mutex
	limiters    map[string]*ipLimiter
	rps         rate.Limit
	burst       int
	proxies     []netip.Prefix
	nextCleanup time.Time
}

func NewRateLimiter(rps float64, burst int, proxies ...netip.Prefix) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*ipLimiter),
		rps:      rate.Limit(rps),
		burst:    burst,
		proxies:  append([]netip.Prefix(nil), proxies...),
	}
	if os.Getenv("ENV") == "test" && os.Getenv("TEST_DISABLE_RATE_LIMIT") == "1" {
		rl.rps = 0
	}
	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.rps <= 0 {
		return nil
	}

	now := time.Now()
	if now.After(rl.nextCleanup) {
		for key, entry := range rl.limiters {
			if now.Sub(entry.lastSeen) > 10*time.Minute {
				delete(rl.limiters, key)
			}
		}
		rl.nextCleanup = now.Add(time.Minute)
	}
	entry, exists := rl.limiters[ip]
	if !exists {
		if len(rl.limiters) >= 10000 {
			return rate.NewLimiter(0, 0)
		}
		limiter := rate.NewLimiter(rl.rps, rl.burst)
		rl.limiters[ip] = &ipLimiter{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := rl.getLimiter(clientIP(r, rl.proxies))
		if limiter != nil {
			if !limiter.Allow() {
				w.Header().Set("Retry-After", "2")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"too many requests, please try again later"}`))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
