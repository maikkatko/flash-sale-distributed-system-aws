package inventory

import (
	"database/sql"
	"fmt"

	"orders/pkg/models"
)

// PessimisticStrategy uses SELECT FOR UPDATE database locking
type PessimisticStrategy struct{}

func (s *PessimisticStrategy) GetName() string {
	return "pessimistic"
}

func (s *PessimisticStrategy) CheckAndReserve(tx *sql.Tx, product *models.Product, quantity int) error {
	// Product is already locked via SELECT FOR UPDATE in the transaction
	// Just validate stock
	if product.Stock < quantity {
		return fmt.Errorf("insufficient stock: only %d available", product.Stock)
	}

	// Stock check passed - reservation implicit via database lock
	return nil
}
