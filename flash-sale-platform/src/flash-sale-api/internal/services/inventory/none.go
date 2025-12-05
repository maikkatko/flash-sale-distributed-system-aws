package inventory

import (
	"context"

	"flash-sale-api/internal/repository"

	"github.com/google/uuid"
)

// NoLockStrategy implements Strategy with no locking
// WARNING: This WILL cause overselling under contention - use only for baseline testing
type NoLockStrategy struct {
	redis *repository.RedisRepository
}

func NewNoLockStrategy(redis *repository.RedisRepository) *NoLockStrategy {
	return &NoLockStrategy{redis: redis}
}

func (s *NoLockStrategy) Name() string {
	return "none"
}

func (s *NoLockStrategy) AttemptPurchase(ctx context.Context, req PurchaseRequest, price float64) (*Reservation, error) {
	// Direct decrement without any locking
	// Race condition: multiple goroutines can read same value before any decrement
	newStock, err := s.redis.DecrInventory(ctx, req.ProductID, req.Quantity)
	if err != nil {
		return nil, err
	}

	// Check if we went negative (oversold)
	if newStock < 0 {
		// Rollback
		s.redis.IncrInventory(ctx, req.ProductID, req.Quantity)
		return nil, NewInsufficientInventoryError(int(newStock + int64(req.Quantity)))
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

func (s *NoLockStrategy) ReleasePurchase(ctx context.Context, productID int, quantity int) error {
	return s.redis.IncrInventory(ctx, productID, quantity)
}
