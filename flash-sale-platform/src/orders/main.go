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
	"github.com/redis/go-redis/v9"
)

var (
	db          *sql.DB
	redisClient *redis.Client
	sqsClient   *sqs.Client
	ctx         = context.Background()
)

// Incoming request structure for purchasing a product
type PurchaseRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	ProductID int    `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// Product represents product data
type Product struct {
	ID    int
	Price float64
	Stock int
}

// OrderMessage published to SQS
type OrderMessage struct {
	UserID     string  `json:"user_id"`
	ProductID  int     `json:"product_id"`
	Quantity   int     `json:"quantity"`
	TotalPrice float64 `json:"total_price"`
	Timestamp  string  `json:"timestamp"`
}

// Initialize database connection
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

	// Retry connection to handle slow database startup
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

	// Optimize connection pool for performance
	db.SetMaxOpenConns(100) // Max number of open connections to the database
	db.SetMaxIdleConns(10)  // Max number of connections in the idle connection pool
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Orders Service: Successfully connected to the database.")
}

// Initialize AWS SQS client
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

// Initialize Redis client
func initRedis() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR environment variable is not set")
	}
	redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	//Test Redis connection
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Orders Service: Successfully connected to Redis.")
}

// acquireLock tries to acquire a distributed lock using Redis
func acquireLock(productID int, timeout time.Duration) (bool, error) {
	lockKey := fmt.Sprintf("lock:product:%d", productID)
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())

	//Acquire lock with expiration
	acquired, err := redisClient.SetNX(ctx, lockKey, lockValue, timeout).Result()
	if err != nil {
		return false, err
	}
	return acquired, nil
}

// releaseLock releases the distributed lock in Redis
func releaseLock(productID int) error {
	lockKey := fmt.Sprintf("lock:product:%d", productID)
	return redisClient.Del(ctx, lockKey).Err()
}

// publishOrderMessage publishes order message to SQS
func publishOrderMessage(msg OrderMessage) error {
	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		return fmt.Errorf("SQS_QUEUE_URL is not set")
	}

	//Marshal message to JSON
	messageBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal order message: %v", err)
	}

	//Send to SQS
	_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(messageBody)),
	})
	if err != nil {
		return fmt.Errorf("failed to send message to SQS: %v", err)
	}
	log.Printf("Published order message to SQS: %+v", msg)
	return nil
}

// purchaseProduct handles the purchase transaction
func purchaseProduct(c *gin.Context) {
	//Parse request body
	var req PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	log.Printf("Purchase request: User %s, Product %d, Quantity %d",
		req.UserID, req.ProductID, req.Quantity)

	// Acquire distributed lock
	lockTimeout := 5 * time.Second
	acquired, err := acquireLock(req.ProductID, lockTimeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to acquire lock"})
		log.Printf("Failed to acquire lock for product %d: %v", req.ProductID, err)
		return
	}
	if !acquired {
		c.JSON(http.StatusConflict, gin.H{"error": "Could not acquire lock, please try again"})
		log.Printf("Could not acquire lock for product %d", req.ProductID)
		return
	}
	defer releaseLock(req.ProductID) // Ensure lock is released
	log.Printf("Acquired lock for product %d", req.ProductID)

	// Start database transaction
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback() // Rollback if not committed

	//Lock product row and check stock
	var product Product
	err = tx.QueryRow(
		"SELECT id, price, stock FROM products WHERE id = ? FOR UPDATE",
		req.ProductID,
	).Scan(&product.ID, &product.Price, &product.Stock)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		log.Printf("Database error: %v", err)
		return
	}

	//Check stock availability
	if product.Stock < req.Quantity {
		c.JSON(http.StatusConflict, gin.H{
			"error": fmt.Sprintf("Insufficient stock: only %d available", product.Stock),
		})
		return
	}

	//Decrement stock
	_, err = tx.Exec(
		"UPDATE products SET stock = stock - ? WHERE id = ?",
		req.Quantity, req.ProductID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock"})
		log.Printf("Failed to update stock: %v", err)
		return
	}

	log.Printf("Stock decremented: Product %d, New stock: %d",
		req.ProductID, product.Stock-req.Quantity)

	//Commit transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		log.Printf("Transaction commit failed: %v", err)
		return
	}

	//Publish order message to SQS
	totalPrice := product.Price * float64(req.Quantity)
	OrderMessage := OrderMessage{
		UserID:     req.UserID,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		TotalPrice: totalPrice,
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	if err := publishOrderMessage(OrderMessage); err != nil {
		//Order not published, but stock is decremented
		//In production, retry logic or Dead Letter Queue
		log.Printf("SQS publish failed (Stock already decremented): %v", err)
		c.JSON(http.StatusAccepted, gin.H{
			"message":     "Purchase accepted but order processing delayed",
			"user_id":     req.UserID,
			"product_id":  req.ProductID,
			"quantity":    req.Quantity,
			"total_price": totalPrice,
		})
		return
	}

	//Return success
	c.JSON(http.StatusCreated, gin.H{
		"user_id":     req.UserID,
		"product_id":  req.ProductID,
		"quantity":    req.Quantity,
		"total_price": totalPrice,
		"message":     "Purchase successful - order queued for processing",
	})

	log.Printf("Purchase successful:User %s, Product %d, Quantity %d, Total Price %.2f",
		req.UserID, req.ProductID, req.Quantity, totalPrice)
}

func main() {
	//Initialize connections
	initDB()
	defer db.Close()

	initSQS()
	initRedis()
	defer redisClient.Close()

	//Setup Gin router
	router := gin.Default()

	//Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "Orders Service is healthy"})
	})

	//Purchase endpoint
	router.POST("/purchase", purchaseProduct)

	//Get port from environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	log.Printf("Orders Service is running on port %s", port)
	router.Run(":" + port)
}
