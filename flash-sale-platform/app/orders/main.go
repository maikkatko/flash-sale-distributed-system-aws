package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	db            *sql.DB
	redisClient   *redis.Client
	sqsClient     *sqs.Client
	ctx           = context.Background()
	reserveScript *redis.Script
	releaseScript *redis.Script
)

// Lua script for atomic inventory reservation
// Returns: {status_code, message, remaining_inventory}
// status_code: 1 = success, 0 = failure
const reserveInventoryLua = `
local inv_key = KEYS[1]           -- inv:{product_id}
local res_key = KEYS[2]           -- res:{order_id}
local quantity = tonumber(ARGV[1])
local order_id = ARGV[2]
local user_id = ARGV[3]
local product_id = ARGV[4]
local ttl = tonumber(ARGV[5])     -- reservation TTL (300s default)

-- Check current inventory
local current = tonumber(redis.call('GET', inv_key) or 0)
if current < quantity then
    return {0, 'INSUFFICIENT_INVENTORY', current}
end

-- Atomic decrement and create reservation
redis.call('DECRBY', inv_key, quantity)
redis.call('HSET', res_key, 
    'qty', quantity, 
    'user_id', user_id, 
    'product_id', product_id,
    'created_at', redis.call('TIME')[1])
redis.call('EXPIRE', res_key, ttl)

return {1, 'RESERVED', current - quantity}
`

// Lua script for releasing reservation (on failure/timeout)
const releaseInventoryLua = `
local inv_key = KEYS[1]           -- inv:{product_id}
local res_key = KEYS[2]           -- res:{order_id}

-- Check if reservation exists
local qty = redis.call('HGET', res_key, 'qty')
if not qty then
    return {0, 'RESERVATION_NOT_FOUND', 0}
end

-- Release inventory and delete reservation
redis.call('INCRBY', inv_key, tonumber(qty))
redis.call('DEL', res_key)

local new_inv = redis.call('GET', inv_key)
return {1, 'RELEASED', tonumber(new_inv)}
`

// PurchaseRequest - incoming request structure
type PurchaseRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	ProductID int    `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// Product represents product data from DB
type Product struct {
	ID    int
	Price float64
	Stock int
}

// OrderMessage published to SQS
type OrderMessage struct {
	OrderID    string  `json:"order_id"`
	UserID     string  `json:"user_id"`
	ProductID  int     `json:"product_id"`
	Quantity   int     `json:"quantity"`
	TotalPrice float64 `json:"total_price"`
	Timestamp  string  `json:"timestamp"`
}

// ReservationResult from Lua script
type ReservationResult struct {
	Success   bool
	Message   string
	Remaining int64
}

func initDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"))

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Failed to ping database, retrying in 5s... (%v)", err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to database after retries: %v", err)
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Orders Service: Successfully connected to the database.")
}

func initSQS() {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	sqsClient = sqs.NewFromConfig(cfg)
	log.Println("Orders Service: Initialized SQS client.")
}

func initRedis() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR environment variable is not set")
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           0,
		PoolSize:     50,
		MinIdleConns: 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  50 * time.Millisecond, // Fast timeout for flash sale
		WriteTimeout: 50 * time.Millisecond,
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Pre-load Lua scripts
	reserveScript = redis.NewScript(reserveInventoryLua)
	releaseScript = redis.NewScript(releaseInventoryLua)

	log.Println("Orders Service: Successfully connected to Redis and loaded Lua scripts.")
}

// reserveInventory uses atomic Lua script to reserve inventory
func reserveInventory(productID int, quantity int, orderID, userID string) (*ReservationResult, error) {
	invKey := fmt.Sprintf("inv:%d", productID)
	resKey := fmt.Sprintf("res:%s", orderID)
	ttl := 300 // 5 minute reservation TTL

	result, err := reserveScript.Run(ctx, redisClient,
		[]string{invKey, resKey},
		quantity, orderID, userID, productID, ttl,
	).Slice()

	if err != nil {
		return nil, fmt.Errorf("lua script error: %v", err)
	}

	statusCode, _ := result[0].(int64)
	message, _ := result[1].(string)
	remaining, _ := result[2].(int64)

	return &ReservationResult{
		Success:   statusCode == 1,
		Message:   message,
		Remaining: remaining,
	}, nil
}

// releaseReservation releases inventory back if order fails
func releaseReservation(productID int, orderID string) error {
	invKey := fmt.Sprintf("inv:%d", productID)
	resKey := fmt.Sprintf("res:%s", orderID)

	_, err := releaseScript.Run(ctx, redisClient,
		[]string{invKey, resKey},
	).Slice()

	return err
}

// syncInventoryToRedis loads inventory from DB to Redis (call on startup or reconciliation)
func syncInventoryToRedis(productID int) error {
	var stock int
	err := db.QueryRow("SELECT stock FROM products WHERE id = ?", productID).Scan(&stock)
	if err != nil {
		return err
	}

	invKey := fmt.Sprintf("inv:%d", productID)
	return redisClient.Set(ctx, invKey, stock, 0).Err()
}

// getProductCached fetches product from Redis cache, falls back to DB
func getProductCached(productID int) (*Product, error) {
	prodKey := fmt.Sprintf("prod:%d", productID)

	// Try Redis first
	result, err := redisClient.HGetAll(ctx, prodKey).Result()
	if err == nil && len(result) > 0 {
		var price float64
		var stock int
		fmt.Sscanf(result["price"], "%f", &price)
		fmt.Sscanf(result["stock"], "%d", &stock)
		return &Product{ID: productID, Price: price, Stock: stock}, nil
	}

	// Cache miss - fetch from DB
	var product Product
	err = db.QueryRow(
		"SELECT id, price, stock FROM products WHERE id = ?",
		productID,
	).Scan(&product.ID, &product.Price, &product.Stock)
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes (price doesn't change often during sale)
	redisClient.HSet(ctx, prodKey,
		"price", fmt.Sprintf("%.2f", product.Price),
		"stock", product.Stock,
	)
	redisClient.Expire(ctx, prodKey, 5*time.Minute)

	return &product, nil
}

func publishOrderMessage(msg OrderMessage) error {
	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		return fmt.Errorf("SQS_QUEUE_URL is not set")
	}

	messageBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal order message: %v", err)
	}

	_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(messageBody)),
	})
	if err != nil {
		return fmt.Errorf("failed to send message to SQS: %v", err)
	}

	log.Printf("Published order message to SQS: %s", msg.OrderID)
	return nil
}

// purchaseProduct - Phase 1: Synchronous reservation via Redis Lua script
func purchaseProduct(c *gin.Context) {
	var req PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	orderID := uuid.New().String()
	log.Printf("[%s] Purchase request: User %s, Product %d, Quantity %d",
		orderID, req.UserID, req.ProductID, req.Quantity)

	// Get product price (Redis cache with DB fallback)
	product, err := getProductCached(req.ProductID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
		log.Printf("[%s] Product fetch error: %v", orderID, err)
		return
	}

	// Phase 1: Atomic inventory reservation via Lua script
	result, err := reserveInventory(req.ProductID, req.Quantity, orderID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Reservation system error"})
		log.Printf("[%s] Reservation error: %v", orderID, err)
		return
	}

	if !result.Success {
		c.JSON(http.StatusConflict, gin.H{
			"error":     result.Message,
			"remaining": result.Remaining,
		})
		log.Printf("[%s] Reservation failed: %s (remaining: %d)", orderID, result.Message, result.Remaining)
		return
	}

	log.Printf("[%s] Inventory reserved. Remaining: %d", orderID, result.Remaining)

	// Phase 2: Queue order for async processing
	totalPrice := product.Price * float64(req.Quantity)
	orderMsg := OrderMessage{
		OrderID:    orderID,
		UserID:     req.UserID,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		TotalPrice: totalPrice,
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	if err := publishOrderMessage(orderMsg); err != nil {
		// SQS failed - release reservation
		if releaseErr := releaseReservation(req.ProductID, orderID); releaseErr != nil {
			log.Printf("[%s] CRITICAL: Failed to release reservation after SQS error: %v", orderID, releaseErr)
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Order queue unavailable, please retry"})
		log.Printf("[%s] SQS publish failed: %v", orderID, err)
		return
	}

	// Return 202 Accepted - order is queued for processing
	c.JSON(http.StatusAccepted, gin.H{
		"order_id":    orderID,
		"status":      "RESERVED",
		"message":     "Order reserved and queued for processing",
		"product_id":  req.ProductID,
		"quantity":    req.Quantity,
		"total_price": totalPrice,
	})

	log.Printf("[%s] Purchase accepted - queued for processing", orderID)
}

// initInventory endpoint to sync a product's inventory from DB to Redis
func initInventory(c *gin.Context) {
	productID := c.Param("id")
	var id int
	fmt.Sscanf(productID, "%d", &id)

	if err := syncInventoryToRedis(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync inventory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inventory synced to Redis", "product_id": id})
}

// getInventory returns current Redis inventory for a product
func getInventory(c *gin.Context) {
	productID := c.Param("id")
	invKey := fmt.Sprintf("inv:%s", productID)

	stock, err := redisClient.Get(ctx, invKey).Int64()
	if err == redis.Nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory not initialized in Redis"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product_id": productID, "available": stock})
}

func main() {
	initDB()
	defer db.Close()

	initSQS()
	initRedis()
	defer redisClient.Close()

	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "Orders Service is healthy"})
	})

	// Purchase endpoint (uses Lua script)
	router.POST("/purchase", purchaseProduct)

	// Inventory management endpoints
	router.POST("/inventory/:id/init", initInventory) // Sync DB -> Redis
	router.GET("/inventory/:id", getInventory)        // Check Redis inventory

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Orders Service is running on port %s", port)
	router.Run(":" + port)
}
