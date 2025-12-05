package inventory

import (
	"context"
)

// PurchaseRequest contains the data needed for a purchase attempt
type PurchaseRequest struct {
	UserID         string
	ProductID      int
	Quantity       int
	IdempotencyKey string
}

// Reservation represents a successful inventory reservation
type Reservation struct {
	OrderID   string
	UserID    string
	ProductID int
	Quantity  int
	Price     float64
	Total     float64
}

// Strategy defines the interface for inventory management strategies
type Strategy interface {
	// Name returns the strategy identifier for logging/metrics
	Name() string

	// AttemptPurchase tries to reserve inventory atomically
	// Returns a Reservation on success, error on failure
	AttemptPurchase(ctx context.Context, req PurchaseRequest, price float64) (*Reservation, error)

	// ReleasePurchase releases a reservation (rollback)
	ReleasePurchase(ctx context.Context, productID int, quantity int) error
}

// Common errors
type InventoryError struct {
	Code      string
	Message   string
	Available int
}

func (e *InventoryError) Error() string {
	return e.Message
}

func NewInsufficientInventoryError(available int) *InventoryError {
	return &InventoryError{
		Code:      "INSUFFICIENT_INVENTORY",
		Message:   "insufficient inventory",
		Available: available,
	}
}

func NewLockError(msg string) *InventoryError {
	return &InventoryError{
		Code:    "LOCK_ERROR",
		Message: msg,
	}
}
