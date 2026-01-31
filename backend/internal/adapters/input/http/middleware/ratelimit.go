package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	limit    int           // requests per window
	window   time.Duration // time window
}

type visitor struct {
	tokens     int
	lastAccess time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}

	// Clean up old visitors every minute
	go rl.cleanupLoop()

	return rl
}

func (rl *RateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			tokens:     rl.limit,
			lastAccess: time.Now(),
		}
		rl.visitors[ip] = v
	}

	return v
}

func (rl *RateLimiter) Allow(ip string) bool {
	v := rl.getVisitor(ip)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(v.lastAccess)

	// Refill tokens based on elapsed time
	if elapsed > rl.window {
		v.tokens = rl.limit
	} else {
		// Partial refill
		refill := int(elapsed.Seconds() / rl.window.Seconds() * float64(rl.limit))
		v.tokens = min(rl.limit, v.tokens+refill)
	}

	v.lastAccess = now

	// Try to consume a token
	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastAccess) > rl.window*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware function
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get IP from X-Forwarded-For or RemoteAddr
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		}

		if !rl.Allow(ip) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
