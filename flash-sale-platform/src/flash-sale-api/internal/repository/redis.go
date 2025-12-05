package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

// Inventory operations
func (r *RedisRepository) GetInventory(ctx context.Context, productID int) (int, error) {
	key := fmt.Sprintf("inv:%d", productID)
	val, err := r.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get inventory: %w", err)
	}
	return val, nil
}

func (r *RedisRepository) SetInventory(ctx context.Context, productID int, stock int) error {
	key := fmt.Sprintf("inv:%d", productID)
	return r.client.Set(ctx, key, stock, 0).Err()
}

func (r *RedisRepository) DecrInventory(ctx context.Context, productID int, quantity int) (int64, error) {
	key := fmt.Sprintf("inv:%d", productID)
	return r.client.DecrBy(ctx, key, int64(quantity)).Result()
}

func (r *RedisRepository) IncrInventory(ctx context.Context, productID int, quantity int) error {
	key := fmt.Sprintf("inv:%d", productID)
	return r.client.IncrBy(ctx, key, int64(quantity)).Err()
}

func (r *RedisRepository) DeleteInventory(ctx context.Context, productID int) error {
	key := fmt.Sprintf("inv:%d", productID)
	return r.client.Del(ctx, key).Err()
}

// Lock operations
func (r *RedisRepository) AcquireLock(ctx context.Context, productID int, value string, ttl time.Duration) (bool, error) {
	key := fmt.Sprintf("lock:%d", productID)
	return r.client.SetNX(ctx, key, value, ttl).Result()
}

func (r *RedisRepository) ReleaseLock(ctx context.Context, productID int, value string) error {
	key := fmt.Sprintf("lock:%d", productID)
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	return r.client.Eval(ctx, script, []string{key}, value).Err()
}

// Idempotency operations
func (r *RedisRepository) GetIdempotencyKey(ctx context.Context, key string) (string, error) {
	idemKey := fmt.Sprintf("idem:%s", key)
	val, err := r.client.Get(ctx, idemKey).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get idempotency: %w", err)
	}
	return val, nil
}

func (r *RedisRepository) SetIdempotencyKey(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	idemKey := fmt.Sprintf("idem:%s", key)
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal idempotency value: %w", err)
	}
	return r.client.Set(ctx, idemKey, string(data), ttl).Err()
}

// Queue operations (for queue-based strategy)
func (r *RedisRepository) EnqueueUser(ctx context.Context, productID int, userID string) (int64, error) {
	key := fmt.Sprintf("queue:%d", productID)
	return r.client.RPush(ctx, key, userID).Result()
}

func (r *RedisRepository) DequeueUser(ctx context.Context, productID int) (string, error) {
	key := fmt.Sprintf("queue:%d", productID)
	return r.client.LPop(ctx, key).Result()
}

func (r *RedisRepository) GetQueueFront(ctx context.Context, productID int) (string, error) {
	key := fmt.Sprintf("queue:%d", productID)
	val, err := r.client.LIndex(ctx, key, 0).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (r *RedisRepository) RemoveFromQueue(ctx context.Context, productID int, userID string) error {
	key := fmt.Sprintf("queue:%d", productID)
	return r.client.LRem(ctx, key, 1, userID).Err()
}

// Lua script execution
func (r *RedisRepository) LoadScript(ctx context.Context, script string) (string, error) {
	return r.client.ScriptLoad(ctx, script).Result()
}

func (r *RedisRepository) EvalSha(ctx context.Context, sha string, keys []string, args ...interface{}) (interface{}, error) {
	return r.client.EvalSha(ctx, sha, keys, args...).Result()
}

// Health check
func (r *RedisRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Client getter for direct access when needed
func (r *RedisRepository) Client() *redis.Client {
	return r.client
}
