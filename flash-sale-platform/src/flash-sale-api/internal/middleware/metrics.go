package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Simple in-memory metrics (replace with Prometheus in production)
type Metrics struct {
	mu             sync.RWMutex
	requestCount   map[string]int64
	requestLatency map[string][]time.Duration
	errorCount     map[string]int64
}

var metrics = &Metrics{
	requestCount:   make(map[string]int64),
	requestLatency: make(map[string][]time.Duration),
	errorCount:     make(map[string]int64),
}

// MetricsMiddleware collects request metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		status := c.Writer.Status()

		metrics.mu.Lock()
		defer metrics.mu.Unlock()

		key := c.Request.Method + " " + path
		metrics.requestCount[key]++
		metrics.requestLatency[key] = append(metrics.requestLatency[key], latency)

		// Keep only last 1000 latencies per endpoint
		if len(metrics.requestLatency[key]) > 1000 {
			metrics.requestLatency[key] = metrics.requestLatency[key][1:]
		}

		if status >= 400 {
			metrics.errorCount[key]++
		}
	}
}

// GetMetrics returns current metrics as a handler
func GetMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics.mu.RLock()
		defer metrics.mu.RUnlock()

		result := make(map[string]interface{})

		for key, count := range metrics.requestCount {
			latencies := metrics.requestLatency[key]
			var p50, p95, p99 time.Duration

			if len(latencies) > 0 {
				sorted := make([]time.Duration, len(latencies))
				copy(sorted, latencies)
				// Simple sort for percentiles
				for i := 0; i < len(sorted); i++ {
					for j := i + 1; j < len(sorted); j++ {
						if sorted[i] > sorted[j] {
							sorted[i], sorted[j] = sorted[j], sorted[i]
						}
					}
				}

				p50 = sorted[len(sorted)*50/100]
				p95 = sorted[len(sorted)*95/100]
				if len(sorted) > 0 {
					p99 = sorted[len(sorted)*99/100]
				}
			}

			result[key] = map[string]interface{}{
				"count":     count,
				"errors":    metrics.errorCount[key],
				"p50_ms":    p50.Milliseconds(),
				"p95_ms":    p95.Milliseconds(),
				"p99_ms":    p99.Milliseconds(),
				"error_pct": strconv.FormatFloat(float64(metrics.errorCount[key])/float64(count)*100, 'f', 2, 64) + "%",
			}
		}

		c.JSON(200, result)
	}
}
