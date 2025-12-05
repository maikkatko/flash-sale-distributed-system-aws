package processor

import (
	"context"
	"fmt"
	"log"
	"time"

	"order-processor/internal/repository"
)

// ProcessRequest contains order data to process
type ProcessRequest struct {
	OrderID        string
	UserID         string
	ProductID      int
	Quantity       int
	TotalPrice     float64
	IdempotencyKey string
	CorrelationID  string
}

// OrderProcessor handles order processing logic
type OrderProcessor struct {
	orderRepo   *repository.OrderRepository
	redisRepo   *repository.RedisRepository
	paymentProc *PaymentProcessor
}

// NewOrderProcessor creates a new order processor
func NewOrderProcessor(
	orderRepo *repository.OrderRepository,
	redisRepo *repository.RedisRepository,
	paymentProc *PaymentProcessor,
) *OrderProcessor {
	return &OrderProcessor{
		orderRepo:   orderRepo,
		redisRepo:   redisRepo,
		paymentProc: paymentProc,
	}
}

// Process handles the complete order processing flow
func (p *OrderProcessor) Process(ctx context.Context, req ProcessRequest) error {
	startTime := time.Now()
	log.Printf("[%s] Starting order processing for user %s, product %d",
		req.CorrelationID, req.UserID, req.ProductID)

	// Step 1: Check idempotency
	exists, err := p.redisRepo.CheckIdempotency(ctx, req.IdempotencyKey)
	if err != nil {
		log.Printf("[%s] Idempotency check failed: %v", req.CorrelationID, err)
		// Continue processing - don't fail on idempotency check error
	}
	if exists {
		log.Printf("[%s] Order already processed (idempotent)", req.CorrelationID)
		return nil
	}

	// Step 2: Simulate payment processing
	paymentResult, err := p.paymentProc.ProcessPayment(ctx, PaymentRequest{
		OrderID:   req.OrderID,
		UserID:    req.UserID,
		Amount:    req.TotalPrice,
		ProductID: req.ProductID,
	})
	if err != nil {
		log.Printf("[%s] Payment failed: %v", req.CorrelationID, err)
		// Release the reservation on payment failure
		if releaseErr := p.releaseReservation(ctx, req); releaseErr != nil {
			log.Printf("[%s] Failed to release reservation: %v", req.CorrelationID, releaseErr)
		}
		return fmt.Errorf("payment failed: %w", err)
	}

	if !paymentResult.Success {
		log.Printf("[%s] Payment declined: %s", req.CorrelationID, paymentResult.Reason)
		if releaseErr := p.releaseReservation(ctx, req); releaseErr != nil {
			log.Printf("[%s] Failed to release reservation: %v", req.CorrelationID, releaseErr)
		}
		return fmt.Errorf("payment declined: %s", paymentResult.Reason)
	}

	// Step 3: Persist order to database
	order := repository.Order{
		ID:            req.OrderID,
		UserID:        req.UserID,
		ProductID:     req.ProductID,
		Quantity:      req.Quantity,
		TotalPrice:    req.TotalPrice,
		Status:        "completed",
		CorrelationID: req.CorrelationID,
	}

	if err := p.orderRepo.CreateOrder(ctx, order); err != nil {
		log.Printf("[%s] Failed to persist order: %v", req.CorrelationID, err)
		// TODO: In production, might need compensation logic for payment
		return fmt.Errorf("failed to persist order: %w", err)
	}

	// Step 4: Update product stock in database
	if err := p.orderRepo.DecrementStock(ctx, req.ProductID, req.Quantity); err != nil {
		log.Printf("[%s] Failed to decrement stock: %v", req.CorrelationID, err)
		// Order is created but stock update failed - needs reconciliation
	}

	// Step 5: Cleanup - delete reservation from Redis
	if err := p.redisRepo.DeleteReservation(ctx, req.OrderID); err != nil {
		log.Printf("[%s] Failed to delete reservation: %v", req.CorrelationID, err)
		// Non-critical - reservation will expire via TTL
	}

	// Step 6: Set idempotency key
	if err := p.redisRepo.SetIdempotency(ctx, req.IdempotencyKey, "completed"); err != nil {
		log.Printf("[%s] Failed to set idempotency key: %v", req.CorrelationID, err)
		// Non-critical
	}

	duration := time.Since(startTime)
	log.Printf("[%s] Order %s completed successfully in %v", req.CorrelationID, req.OrderID, duration)

	return nil
}

// releaseReservation returns inventory to Redis when order fails
func (p *OrderProcessor) releaseReservation(ctx context.Context, req ProcessRequest) error {
	log.Printf("[%s] Releasing reservation for product %d, quantity %d",
		req.CorrelationID, req.ProductID, req.Quantity)

	// Increment inventory counter back
	if err := p.redisRepo.IncrementInventory(ctx, req.ProductID, req.Quantity); err != nil {
		return err
	}

	// Delete the reservation record
	return p.redisRepo.DeleteReservation(ctx, req.OrderID)
}
