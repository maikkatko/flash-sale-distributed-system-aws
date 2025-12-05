package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"flash-sale-api/internal/repository"
	"flash-sale-api/internal/services/inventory"
	"flash-sale-api/internal/services/order"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PurchaseRequest struct {
	UserID         string `json:"user_id" binding:"required"`
	ProductID      int    `json:"product_id" binding:"required"`
	Quantity       int    `json:"quantity" binding:"required,min=1"`
	IdempotencyKey string `json:"idempotency_key"`
}

type PurchaseResponse struct {
	OrderID        string  `json:"order_id"`
	UserID         string  `json:"user_id"`
	ProductID      int     `json:"product_id"`
	Quantity       int     `json:"quantity"`
	TotalPrice     float64 `json:"total_price"`
	Status         string  `json:"status"`
	Message        string  `json:"message,omitempty"`
	IdempotencyKey string  `json:"idempotency_key,omitempty"`
}

type PurchaseHandler struct {
	productRepo *repository.ProductRepository
	redisRepo   *repository.RedisRepository
	strategy    inventory.Strategy
	publisher   *order.Publisher
}

func NewPurchaseHandler(
	productRepo *repository.ProductRepository,
	redisRepo *repository.RedisRepository,
	strategy inventory.Strategy,
	publisher *order.Publisher,
) *PurchaseHandler {
	return &PurchaseHandler{
		productRepo: productRepo,
		redisRepo:   redisRepo,
		strategy:    strategy,
		publisher:   publisher,
	}
}

// Purchase handles the checkout flow
func (h *PurchaseHandler) Purchase(c *gin.Context) {
	ctx := c.Request.Context()

	var req PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate idempotency key if not provided
	if req.IdempotencyKey == "" {
		req.IdempotencyKey = uuid.New().String()
	}

	// Check idempotency - return cached response if duplicate
	cached, err := h.redisRepo.GetIdempotencyKey(ctx, req.IdempotencyKey)
	if err == nil && cached != "" {
		var resp PurchaseResponse
		if json.Unmarshal([]byte(cached), &resp) == nil {
			resp.Message = "duplicate request"
			c.JSON(http.StatusOK, resp)
			return
		}
	}

	// Get product price
	price, err := h.productRepo.GetPrice(ctx, req.ProductID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	// Attempt purchase using configured strategy
	reservation, err := h.strategy.AttemptPurchase(ctx, inventory.PurchaseRequest{
		UserID:         req.UserID,
		ProductID:      req.ProductID,
		Quantity:       req.Quantity,
		IdempotencyKey: req.IdempotencyKey,
	}, price)

	if err != nil {
		// Check if it's an inventory error
		if invErr, ok := err.(*inventory.InventoryError); ok {
			c.JSON(http.StatusConflict, gin.H{
				"error":     invErr.Message,
				"code":      invErr.Code,
				"available": invErr.Available,
				"strategy":  h.strategy.Name(),
			})
			return
		}
		c.JSON(http.StatusConflict, gin.H{
			"error":    err.Error(),
			"strategy": h.strategy.Name(),
		})
		return
	}

	// Publish to SQS for async processing
	orderMsg := order.OrderMessage{
		OrderID:        reservation.OrderID,
		UserID:         req.UserID,
		ProductID:      req.ProductID,
		Quantity:       req.Quantity,
		TotalPrice:     reservation.Total,
		IdempotencyKey: req.IdempotencyKey,
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	if err := h.publisher.Publish(ctx, orderMsg); err != nil {
		// Rollback inventory reservation
		h.strategy.ReleasePurchase(ctx, req.ProductID, req.Quantity)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue order"})
		return
	}

	// Build response
	resp := PurchaseResponse{
		OrderID:        reservation.OrderID,
		UserID:         req.UserID,
		ProductID:      req.ProductID,
		Quantity:       req.Quantity,
		TotalPrice:     reservation.Total,
		Status:         "accepted",
		Message:        "order queued for processing",
		IdempotencyKey: req.IdempotencyKey,
	}

	// Cache response for idempotency (24 hours)
	h.redisRepo.SetIdempotencyKey(ctx, req.IdempotencyKey, resp, 24*time.Hour)

	c.JSON(http.StatusAccepted, resp)
}
