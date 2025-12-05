package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Simple token bucket rate limiter
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     int           // tokens per interval
	interval time.Duration // refill interval
	burst    int           // max tokens
}

type bucket struct {
	tokens     int
	lastRefill time.Time
}

// NewRateLimiter creates a rate limiter
// rate: number of requests allowed per interval
// interval: time window for rate limit
// burst: maximum burst size
func NewRateLimiter(rate int, interval time.Duration, burst int) *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*bucket),
		rate:     rate,
		interval: interval,
		burst:    burst,
	}

	// Cleanup old buckets periodically
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-10 * time.Minute)
	for key, b := range rl.buckets {
		if b.lastRefill.Before(cutoff) {
			delete(rl.buckets, key)
		}
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.buckets[key]

	if !exists {
		rl.buckets[key] = &bucket{
			tokens:     rl.burst - 1,
			lastRefill: now,
		}
		return true
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(b.lastRefill)
	tokensToAdd := int(elapsed / rl.interval) * rl.rate
	if tokensToAdd > 0 {
		b.tokens = min(b.tokens+tokensToAdd, rl.burst)
		b.lastRefill = now
	}

	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// RateLimitMiddleware creates a Gin middleware for rate limiting
// Uses user_id from request body or IP as key
func RateLimitMiddleware(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use user_id from header, or fall back to IP
		key := c.GetHeader("X-User-ID")
		if key == "" {
			key = c.ClientIP()
		}

		if !rl.Allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"retry": rl.interval.String(),
			})
			return
		}

		c.Next()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
