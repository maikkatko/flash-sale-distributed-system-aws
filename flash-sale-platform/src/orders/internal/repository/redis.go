package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisRepository(addr, password string) (*RedisRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx := context.Background()

	// Test connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis")

	return &RedisRepository{
		client: client,
		ctx:    ctx,
	}, nil
}

func (r *RedisRepository) Close() error {
	return r.client.Close()
}

// AcquireLock attempts to acquire a distributed lock
func (r *RedisRepository) AcquireLock(productID int, timeout time.Duration) (bool, error) {
	lockKey := fmt.Sprintf("lock:product:%d", productID)
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())

	acquired, err := r.client.SetNX(r.ctx, lockKey, lockValue, timeout).Result()
	if err != nil {
		return false, err
	}

	return acquired, nil
}

// ReleaseLock releases a distributed lock
func (r *RedisRepository) ReleaseLock(productID int) error {
	lockKey := fmt.Sprintf("lock:product:%d", productID)
	return r.client.Del(r.ctx, lockKey).Err()
}

// IncrementCounter increments a Redis counter (for queue-based strategy)
func (r *RedisRepository) IncrementCounter(key string) (int64, error) {
	return r.client.Incr(r.ctx, key).Result()
}

// GetInventoryCache gets cached inventory count
func (r *RedisRepository) GetInventoryCache(productID int) (int, error) {
	key := fmt.Sprintf("inv:%d", productID)
	val, err := r.client.Get(r.ctx, key).Int()
	if err == redis.Nil {
		return 0, fmt.Errorf("inventory not cached")
	}
	return val, err
}

// SetInventoryCache sets cached inventory count
func (r *RedisRepository) SetInventoryCache(productID, stock int, ttl time.Duration) error {
	key := fmt.Sprintf("inv:%d", productID)
	return r.client.Set(r.ctx, key, stock, ttl).Err()
}
