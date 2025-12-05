package inventory

import (
	"database/sql"
	"orders/pkg/models"
)

// Strategy defines the interface for different inventory management approaches
type Strategy interface {
	// CheckAndReserve validates stock and reserves inventory
	// Returns error if insufficient stock or reservation fails
	CheckAndReserve(tx *sql.Tx, product *models.Product, quantity int) error

	// GetName returns the strategy name for logging/metrics
	GetName() string
}

// StrategyFactory creates inventory strategies based on configuration
func NewStrategy(name string) Strategy {
	switch name {
	case "none":
		return &NoneStrategy{}
	case "pessimistic":
		return &PessimisticStrategy{}
	case "optimistic":
		return &OptimisticStrategy{}
	case "queue":
		return &QueueStrategy{}
	default:
		return &PessimisticStrategy{} // Default to safest
	}
}
