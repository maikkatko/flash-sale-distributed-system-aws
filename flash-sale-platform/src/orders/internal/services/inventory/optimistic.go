package inventory

import (
	"database/sql"
	"fmt"
	"log"

	"orders/pkg/models"
)

// OptimisticStrategy uses compare-and-swap without locking
type OptimisticStrategy struct {
	productRepo ProductRepository
}

// ProductRepository interface for database operations
type ProductRepository interface {
	GetProduct(productID int) (*models.Product, error)
	DecrementStockOptimistic(productID, quantity, expectedStock int) (bool, error)
}

func NewOptimisticStrategy(repo ProductRepository) *OptimisticStrategy {
	return &OptimisticStrategy{
		productRepo: repo,
	}
}

func (s *OptimisticStrategy) GetName() string {
	return "optimistic"
}

func (s *OptimisticStrategy) CheckAndReserve(tx *sql.Tx, product *models.Product, quantity int) error {
	// Optimistic approach: No locks, use compare-and-swap

	// Note: We already have the product from the transaction's SELECT FOR UPDATE
	// But for true optimistic locking, we'd read WITHOUT locking first
	// For simplicity in this refactor, we'll validate and rely on the UPDATE with WHERE clause

	if product.Stock < quantity {
		return fmt.Errorf("insufficient stock: only %d available", product.Stock)
	}

	// The actual optimistic check happens in the repository layer
	// UPDATE products SET stock = stock - ? WHERE id = ? AND stock >= ?
	// If stock changed between read and update, this fails (compare-and-swap)

	return nil
}

// CheckAndReserveOptimistic is the true optimistic implementation
// (Called outside of transaction for genuine optimistic concurrency)
func (s *OptimisticStrategy) CheckAndReserveOptimistic(productID, quantity int, maxRetries int) error {
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Read current stock (no lock)
		product, err := s.productRepo.GetProduct(productID)
		if err != nil {
			return err
		}

		// Check if enough stock
		if product.Stock < quantity {
			return fmt.Errorf("insufficient stock: only %d available", product.Stock)
		}

		// Attempt optimistic update (compare-and-swap)
		success, err := s.productRepo.DecrementStockOptimistic(productID, quantity, product.Stock)
		if err != nil {
			return err
		}

		if success {
			// Update succeeded - stock was still at expected value
			log.Printf("Optimistic update succeeded (attempt %d)", attempt+1)
			return nil
		}

		// Stock changed between read and update - retry
		log.Printf("Optimistic conflict (attempt %d/%d) - retrying", attempt+1, maxRetries)
	}

	return fmt.Errorf("optimistic update failed after %d retries (high contention)", maxRetries)
}
