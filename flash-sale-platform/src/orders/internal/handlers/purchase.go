package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"orders/internal/repository"
	"orders/internal/services/inventory"
	"orders/pkg/models"
)

type PurchaseHandler struct {
	productRepo     *repository.ProductRepository
	redisRepo       *repository.RedisRepository
	sqsRepo         *repository.SQSRepository
	lockingStrategy inventory.Strategy
}

func NewPurchaseHandler(
	productRepo *repository.ProductRepository,
	redisRepo *repository.RedisRepository,
	sqsRepo *repository.SQSRepository,
	strategy inventory.Strategy,
) *PurchaseHandler {
	return &PurchaseHandler{
		productRepo:     productRepo,
		redisRepo:       redisRepo,
		sqsRepo:         sqsRepo,
		lockingStrategy: strategy,
	}
}

// Purchase handles POST /purchase requests
func (h *PurchaseHandler) Purchase(c *gin.Context) {
	// Parse request
	var req models.PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	log.Printf("Purchase: User %s, Product %d, Qty %d [Strategy: %s]",
		req.UserID, req.ProductID, req.Quantity, h.lockingStrategy.GetName())

	// Acquire Redis distributed lock (cross-instance coordination)
	lockTimeout := 5 * time.Second
	acquired, err := h.redisRepo.AcquireLock(req.ProductID, lockTimeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lock acquisition failed"})
		log.Printf("Lock error: %v", err)
		return
	}
	if !acquired {
		c.JSON(http.StatusConflict, gin.H{"error": "Product locked, try again"})
		log.Printf("Lock contention for product %d", req.ProductID)
		return
	}
	defer h.redisRepo.ReleaseLock(req.ProductID)
	log.Printf("Lock acquired for product %d", req.ProductID)

	// Start database transaction
	tx, err := h.productRepo.BeginTx()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}
	defer tx.Rollback()

	// Get product with appropriate locking based on strategy
	product, err := h.productRepo.GetProductWithLock(tx, req.ProductID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			log.Printf("DB error: %v", err)
		}
		return
	}

	// Apply locking strategy (validates stock based on strategy)
	if err := h.lockingStrategy.CheckAndReserve(tx, product, req.Quantity); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		log.Printf("Strategy check failed: %v", err)
		return
	}

	// Decrement stock
	if err := h.productRepo.DecrementStock(tx, req.ProductID, req.Quantity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock"})
		log.Printf("Stock update failed: %v", err)
		return
	}

	log.Printf("Stock decremented: Product %d, %d → %d",
		req.ProductID, product.Stock, product.Stock-req.Quantity)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit failed"})
		log.Printf("Commit failed: %v", err)
		return
	}

	// Publish to SQS (async processing)
	totalPrice := product.Price * float64(req.Quantity)
	orderMsg := models.OrderMessage{
		UserID:     req.UserID,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		TotalPrice: totalPrice,
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	// Use retry logic for SQS (handles transient failures)
	if err := h.sqsRepo.PublishOrderWithRetry(orderMsg, 3); err != nil {
		log.Printf("SQS publish failed (stock decremented): %v", err)
		c.JSON(http.StatusAccepted, gin.H{
			"message":     "Purchase accepted but order processing delayed",
			"user_id":     req.UserID,
			"product_id":  req.ProductID,
			"quantity":    req.Quantity,
			"total_price": totalPrice,
		})
		return
	}

	// Success
	c.JSON(http.StatusCreated, gin.H{
		"message":     "Purchase successful - order queued for processing",
		"user_id":     req.UserID,
		"product_id":  req.ProductID,
		"quantity":    req.Quantity,
		"total_price": totalPrice,
	})

	log.Printf("Purchase complete: User %s, Product %d, Qty %d, Price %.2f",
		req.UserID, req.ProductID, req.Quantity, totalPrice)
}
