package inventory

import (
	"context"
	"fmt"
	"time"

	"flash-sale-api/internal/repository"

	"github.com/google/uuid"
)

// QueueStrategy implements Strategy using FIFO queuing
// Users wait in a per-product queue for their turn - fairest but slowest
type QueueStrategy struct {
	redis       *repository.RedisRepository
	waitTimeout time.Duration
	pollInterval time.Duration
}

func NewQueueStrategy(redis *repository.RedisRepository) *QueueStrategy {
	return &QueueStrategy{
		redis:        redis,
		waitTimeout:  10 * time.Second,
		pollInterval: 50 * time.Millisecond,
	}
}

func (s *QueueStrategy) Name() string {
	return "queue"
}

func (s *QueueStrategy) AttemptPurchase(ctx context.Context, req PurchaseRequest, price float64) (*Reservation, error) {
	// Add user to queue
	_, err := s.redis.EnqueueUser(ctx, req.ProductID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("enqueue user: %w", err)
	}

	// Wait for our turn
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	timeout := time.After(s.waitTimeout)

	for {
		select {
		case <-ctx.Done():
			s.redis.RemoveFromQueue(ctx, req.ProductID, req.UserID)
			return nil, ctx.Err()

		case <-timeout:
			s.redis.RemoveFromQueue(ctx, req.ProductID, req.UserID)
			return nil, NewLockError("queue timeout - too many users ahead")

		case <-ticker.C:
			front, err := s.redis.GetQueueFront(ctx, req.ProductID)
			if err != nil {
				continue
			}

			if front == req.UserID {
				// Our turn - process purchase
				reservation, err := s.processInQueue(ctx, req, price)

				// Remove from queue regardless of outcome
				s.redis.DequeueUser(ctx, req.ProductID)

				return reservation, err
			}
		}
	}
}

func (s *QueueStrategy) processInQueue(ctx context.Context, req PurchaseRequest, price float64) (*Reservation, error) {
	// Check inventory
	stock, err := s.redis.GetInventory(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	if stock < req.Quantity {
		return nil, NewInsufficientInventoryError(stock)
	}

	// Decrement (we have exclusive access via queue position)
	_, err = s.redis.DecrInventory(ctx, req.ProductID, req.Quantity)
	if err != nil {
		return nil, err
	}

	return &Reservation{
		OrderID:   uuid.New().String(),
		UserID:    req.UserID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Price:     price,
		Total:     price * float64(req.Quantity),
	}, nil
}

func (s *QueueStrategy) ReleasePurchase(ctx context.Context, productID int, quantity int) error {
	return s.redis.IncrInventory(ctx, productID, quantity)
}
