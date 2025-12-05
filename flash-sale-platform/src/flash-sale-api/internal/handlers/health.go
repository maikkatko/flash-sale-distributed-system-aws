package handlers

import (
	"database/sql"
	"net/http"

	"flash-sale-api/internal/repository"
	"flash-sale-api/internal/services/inventory"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db       *sql.DB
	redis    *repository.RedisRepository
	strategy inventory.Strategy
}

func NewHealthHandler(db *sql.DB, redis *repository.RedisRepository, strategy inventory.Strategy) *HealthHandler {
	return &HealthHandler{
		db:       db,
		redis:    redis,
		strategy: strategy,
	}
}

// Health is a simple liveness check
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"strategy": h.strategy.Name(),
	})
}

// Ready is a readiness check that verifies dependencies
func (h *HealthHandler) Ready(c *gin.Context) {
	// Check database
	if err := h.db.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "database unavailable",
		})
		return
	}

	// Check Redis
	if err := h.redis.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "redis unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "ready",
		"strategy": h.strategy.Name(),
	})
}
