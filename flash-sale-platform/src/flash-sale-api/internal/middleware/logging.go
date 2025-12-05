package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const CorrelationIDHeader = "X-Correlation-ID"

// Logger logs request details with timing
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Generate or extract correlation ID
		correlationID := c.GetHeader(CorrelationIDHeader)
		if correlationID == "" {
			correlationID = uuid.New().String()
		}
		c.Set("correlation_id", correlationID)
		c.Header(CorrelationIDHeader, correlationID)

		// Process request
		c.Next()

		// Log after request
		latency := time.Since(start)
		status := c.Writer.Status()

		log.Printf("[%s] %s %s %d %v",
			correlationID[:8],
			c.Request.Method,
			c.Request.URL.Path,
			status,
			latency,
		)
	}
}
