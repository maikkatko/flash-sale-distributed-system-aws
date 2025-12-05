package inventory

import (
	"context"
	"time"

	"flash-sale-api/internal/repository"

	"github.com/google/uuid"
)

// PessimisticStrategy implements Strategy with distributed locking
// Uses Redis SETNX for mutual exclusion before inventory operations
type PessimisticStrategy struct {
	redis       *repository.RedisRepository
	lockTimeout time.Duration
}

func NewPessimisticStrategy(redis *repository.RedisRepository) *PessimisticStrategy {
	return &PessimisticStrategy{
		redis:       redis,
		lockTimeout: 5 * time.Second,
	}
}

func (s *PessimisticStrategy) Name() string {
	return "pessimistic"
}

func (s *PessimisticStrategy) AttemptPurchase(ctx context.Context, req PurchaseRequest, price float64) (*Reservation, error) {
	lockValue := uuid.New().String()

	// Try to acquire distributed lock
	acquired, err := s.redis.AcquireLock(ctx, req.ProductID, lockValue, s.lockTimeout)
	if err != nil {
		return nil, NewLockError("lock acquisition failed: " + err.Error())
	}
	if !acquired {
		return nil, NewLockError("could not acquire lock - resource busy")
	}

	// Ensure lock is released
	defer s.redis.ReleaseLock(ctx, req.ProductID, lockValue)

	// Check inventory under lock
	stock, err := s.redis.GetInventory(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	if stock < req.Quantity {
		return nil, NewInsufficientInventoryError(stock)
	}

	// Decrement under lock - safe from race conditions
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

func (s *PessimisticStrategy) ReleasePurchase(ctx context.Context, productID int, quantity int) error {
	return s.redis.IncrInventory(ctx, productID, quantity)
}
