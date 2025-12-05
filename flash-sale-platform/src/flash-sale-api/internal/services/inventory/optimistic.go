package inventory

import (
	"context"
	"fmt"

	"flash-sale-api/internal/repository"

	"github.com/google/uuid"
)

// OptimisticStrategy implements Strategy using atomic Lua scripts
// This is the recommended strategy - atomic check-and-decrement prevents overselling
type OptimisticStrategy struct {
	redis     *repository.RedisRepository
	scriptSHA string
}

func NewOptimisticStrategy(ctx context.Context, redis *repository.RedisRepository) (*OptimisticStrategy, error) {
	// Lua script: atomic check-and-decrement
	// Returns: {success (0/1), remaining_stock}
	script := `
		local inv_key = KEYS[1]
		local quantity = tonumber(ARGV[1])
		
		local current = tonumber(redis.call('GET', inv_key) or 0)
		if current < quantity then
			return {0, current}
		end
		
		local new_stock = redis.call('DECRBY', inv_key, quantity)
		return {1, new_stock}
	`

	sha, err := redis.LoadScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("load lua script: %w", err)
	}

	return &OptimisticStrategy{
		redis:     redis,
		scriptSHA: sha,
	}, nil
}

func (s *OptimisticStrategy) Name() string {
	return "optimistic"
}

func (s *OptimisticStrategy) AttemptPurchase(ctx context.Context, req PurchaseRequest, price float64) (*Reservation, error) {
	invKey := fmt.Sprintf("inv:%d", req.ProductID)

	result, err := s.redis.EvalSha(ctx, s.scriptSHA, []string{invKey}, req.Quantity)
	if err != nil {
		return nil, fmt.Errorf("lua script execution: %w", err)
	}

	res := result.([]interface{})
	success := res[0].(int64)
	remaining := res[1].(int64)

	if success == 0 {
		return nil, NewInsufficientInventoryError(int(remaining))
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

func (s *OptimisticStrategy) ReleasePurchase(ctx context.Context, productID int, quantity int) error {
	return s.redis.IncrInventory(ctx, productID, quantity)
}
