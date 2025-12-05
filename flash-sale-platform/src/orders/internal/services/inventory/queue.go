package inventory

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"orders/pkg/models"
)

// QueueStrategy implements FIFO (First-In-First-Out) fairness
// Uses Redis counter to assign queue numbers
type QueueStrategy struct {
	redisRepo RedisRepository
}

// RedisRepository interface for Redis operations
type RedisRepository interface {
	IncrementCounter(key string) (int64, error)
	AcquireLock(productID int, timeout time.Duration) (bool, error)
	ReleaseLock(productID int) error
}

func NewQueueStrategy(redis RedisRepository) *QueueStrategy {
	return &QueueStrategy{
		redisRepo: redis,
	}
}

func (s *QueueStrategy) GetName() string {
	return "queue"
}

func (s *QueueStrategy) CheckAndReserve(tx *sql.Tx, product *models.Product, quantity int) error {
	// Step 1: Get queue number (FIFO ordering)
	queueKey := fmt.Sprintf("queue:product:%d", product.ID)
	queueNumber, err := s.redisRepo.IncrementCounter(queueKey)
	if err != nil {
		return fmt.Errorf("failed to get queue number: %v", err)
	}

	log.Printf("Queue position: %d for product %d", queueNumber, product.ID)

	// Step 2: Acquire lock (ensures FIFO processing)
	lockTimeout := 10 * time.Second
	acquired, err := s.redisRepo.AcquireLock(product.ID, lockTimeout)
	if err != nil {
		return fmt.Errorf("lock acquisition failed: %v", err)
	}
	if !acquired {
		return fmt.Errorf("failed to acquire lock (queue position: %d)", queueNumber)
	}
	defer s.redisRepo.ReleaseLock(product.ID)

	// Step 3: Validate stock (same as other strategies)
	if product.Stock < quantity {
		return fmt.Errorf("insufficient stock: only %d available", product.Stock)
	}

	// Reservation successful with queue fairness guarantee
	return nil
}
