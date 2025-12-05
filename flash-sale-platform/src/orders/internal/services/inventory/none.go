package inventory

import (
	"database/sql"
	"fmt"
	"orders/pkg/models"
)

// NoneStrategy performs no locking - baseline for measuring oversells
type NoneStrategy struct{}

func (s *NoneStrategy) GetName() string {
	return "none"
}

func (s *NoneStrategy) CheckAndReserve(tx *sql.Tx, product *models.Product, quantity int) error {
	// No locking - just check stock (race condition possible!)
	if product.Stock < quantity {
		return ErrInsufficientStock
	}
	return nil
}

var ErrInsufficientStock = fmt.Errorf("insufficient stock")
