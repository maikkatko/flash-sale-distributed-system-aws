package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// Key prefixes
	inventoryKeyPrefix   = "inv:"
	reservationKeyPrefix = "res:"
	idempotencyKeyPrefix = "idem:"

	// TTLs
	idempotencyTTL = 24 * time.Hour
)

// RedisRepository handles Redis operations for inventory and reservations
type RedisRepository struct {
	client *redis.Client
}

// NewRedisRepository creates a new Redis repository
func NewRedisRepository() (*RedisRepository, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		return nil, fmt.Errorf("REDIS_ADDR environment variable is required")
	}

	password := os.Getenv("REDIS_PASSWORD")

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,
		PoolSize:     50,
		MinIdleConns: 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test connection with retries
	var pingErr error
	for i := 0; i < 5; i++ {
		pingErr = client.Ping(context.Background()).Err()
		if pingErr == nil {
			break
		}
		log.Printf("Failed to ping Redis, retrying in 2s... (%v)", pingErr)
		time.Sleep(2 * time.Second)
	}
	if pingErr != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", pingErr)
	}

	log.Println("Redis Repository: Successfully connected to Redis")

	return &RedisRepository{client: client}, nil
}

// CheckIdempotency checks if an idempotency key already exists
func (r *RedisRepository) CheckIdempotency(ctx context.Context, key string) (bool, error) {
	fullKey := idempotencyKeyPrefix + key
	exists, err := r.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check idempotency: %w", err)
	}
	return exists > 0, nil
}

// SetIdempotency sets an idempotency key with result
func (r *RedisRepository) SetIdempotency(ctx context.Context, key string, result string) error {
	fullKey := idempotencyKeyPrefix + key
	err := r.client.SetEx(ctx, fullKey, result, idempotencyTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set idempotency: %w", err)
	}
	return nil
}

// GetIdempotencyResult gets the result for an idempotency key
func (r *RedisRepository) GetIdempotencyResult(ctx context.Context, key string) (string, error) {
	fullKey := idempotencyKeyPrefix + key
	result, err := r.client.Get(ctx, fullKey).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get idempotency result: %w", err)
	}
	return result, nil
}

// DeleteReservation removes a reservation from Redis
func (r *RedisRepository) DeleteReservation(ctx context.Context, orderID string) error {
	fullKey := reservationKeyPrefix + orderID
	err := r.client.Del(ctx, fullKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete reservation: %w", err)
	}
	log.Printf("Deleted reservation: %s", orderID)
	return nil
}

// IncrementInventory increases the inventory counter (for releasing reservations)
func (r *RedisRepository) IncrementInventory(ctx context.Context, productID int, quantity int) error {
	fullKey := fmt.Sprintf("%s%d", inventoryKeyPrefix, productID)
	newVal, err := r.client.IncrBy(ctx, fullKey, int64(quantity)).Result()
	if err != nil {
		return fmt.Errorf("failed to increment inventory: %w", err)
	}
	log.Printf("Inventory incremented for product %d: +%d (new value: %d)", productID, quantity, newVal)
	return nil
}

// GetInventory gets the current inventory count
func (r *RedisRepository) GetInventory(ctx context.Context, productID int) (int, error) {
	fullKey := fmt.Sprintf("%s%d", inventoryKeyPrefix, productID)
	val, err := r.client.Get(ctx, fullKey).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get inventory: %w", err)
	}
	return val, nil
}

// GetReservation gets reservation details
func (r *RedisRepository) GetReservation(ctx context.Context, orderID string) (map[string]string, error) {
	fullKey := reservationKeyPrefix + orderID
	result, err := r.client.HGetAll(ctx, fullKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}

// Close closes the Redis connection
func (r *RedisRepository) Close() error {
	return r.client.Close()
}
