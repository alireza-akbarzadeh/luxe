// Package middleware provides reusable Gin middleware functions, such as rate limiting, logging, and authentication.
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const ipLimiterTTL = 10 * time.Minute

type ipEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter stores a rate limiter per IP address with TTL-based eviction.
type IPRateLimiter struct {
	entries  map[string]*ipEntry
	mu       sync.Mutex
	r        rate.Limit
	b        int
}

// NewIPRateLimiter creates a new rate limiter.
// r: requests per second; b: burst size (max tokens).
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	il := &IPRateLimiter{
		entries: make(map[string]*ipEntry),
		r:       r,
		b:       b,
	}
	go il.cleanupLoop()
	return il
}

func (il *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	il.mu.Lock()
	defer il.mu.Unlock()

	e, exists := il.entries[ip]
	if !exists {
		e = &ipEntry{limiter: rate.NewLimiter(il.r, il.b)}
		il.entries[ip] = e
	}
	e.lastSeen = time.Now()
	return e.limiter
}

// cleanupLoop removes entries that have been idle longer than ipLimiterTTL.
func (il *IPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(ipLimiterTTL)
	defer ticker.Stop()
	for range ticker.C {
		il.mu.Lock()
		cutoff := time.Now().Add(-ipLimiterTTL)
		for ip, e := range il.entries {
			if e.lastSeen.Before(cutoff) {
				delete(il.entries, ip)
			}
		}
		il.mu.Unlock()
	}
}

// RateLimitMiddleware returns a Gin middleware that limits requests per IP.
// Uses utils.Response format so clients see a consistent error envelope.
func RateLimitMiddleware(r rate.Limit, burst int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(r, burst)
	return func(c *gin.Context) {
		if !limiter.getLimiter(c.ClientIP()).Allow() {
			utils.ErrorResponse(c, http.StatusTooManyRequests, "too many requests — slow down and retry")
			c.Abort()
			return
		}
		c.Next()
	}
}

// StrictRateLimit is a convenience preset for sensitive endpoints (auth, import).
// 5 requests per minute (≈0.083 req/s), burst 10.
func StrictRateLimit() gin.HandlerFunc {
	return RateLimitMiddleware(rate.Limit(5.0/60), 10)
}

// StandardRateLimit is the default API-wide preset.
// 100 requests per second, burst 200.
func StandardRateLimit() gin.HandlerFunc {
	return RateLimitMiddleware(100, 200)
}
