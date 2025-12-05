package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	appconfig "flash-sale-api/internal/config"
	"flash-sale-api/internal/handlers"
	"flash-sale-api/internal/middleware"
	"flash-sale-api/internal/repository"
	"flash-sale-api/internal/services/inventory"
	"flash-sale-api/internal/services/order"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load configuration
	cfg := appconfig.Load()

	// Initialize database
	db, err := initDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := initRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize SQS
	sqsClient, err := initSQS(cfg.SQS)
	if err != nil {
		log.Fatalf("Failed to initialize SQS: %v", err)
	}

	// Create repositories
	productRepo := repository.NewProductRepository(db)
	redisRepo := repository.NewRedisRepository(redisClient)

	// Create inventory strategy
	strategy, err := createStrategy(cfg.Strategy, redisRepo)
	if err != nil {
		log.Fatalf("Failed to create inventory strategy: %v", err)
	}

	// Create order publisher
	publisher := order.NewPublisher(sqsClient, cfg.SQS.QueueURL)

	// Create handlers
	healthHandler := handlers.NewHealthHandler(db, redisRepo, strategy)
	productHandler := handlers.NewProductHandler(productRepo, redisRepo)
	purchaseHandler := handlers.NewPurchaseHandler(productRepo, redisRepo, strategy, publisher)
	orderHandler := handlers.NewOrderHandler(db)

	// Setup router
	router := setupRouter(healthHandler, productHandler, purchaseHandler, orderHandler)

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server
	go func() {
		log.Printf("Flash Sale API starting on port %s with strategy %s", cfg.Server.Port, strategy.Name())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func initDB(cfg appconfig.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Retry connection
	for i := 0; i < 10; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		log.Printf("DB connection attempt %d failed: %v", i+1, err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("database connection failed after retries: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Create tables
	if err := createTables(db); err != nil {
		return nil, err
	}

	log.Println("Database initialized")
	return db, nil
}

func createTables(db *sql.DB) error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS products (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price DECIMAL(10, 2) NOT NULL,
			stock INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_name (name)
		) ENGINE=InnoDB;`,

		`CREATE TABLE IF NOT EXISTS orders (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			product_id INT NOT NULL,
			quantity INT NOT NULL DEFAULT 1,
			total_price DECIMAL(10, 2) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			idempotency_key VARCHAR(255) UNIQUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_user (user_id),
			INDEX idx_product (product_id),
			INDEX idx_status (status)
		) ENGINE=InnoDB;`,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return err
		}
	}
	return nil
}

func initRedis(cfg appconfig.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	log.Println("Redis connected")
	return client, nil
}

func initSQS(cfg appconfig.SQSConfig) (*sqs.Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("aws config failed: %w", err)
	}

	log.Println("SQS initialized")
	return sqs.NewFromConfig(awsCfg), nil
}

func createStrategy(name string, redisRepo *repository.RedisRepository) (inventory.Strategy, error) {
	ctx := context.Background()

	switch name {
	case "none":
		log.Println("Using NO LOCK strategy (baseline - will oversell!)")
		return inventory.NewNoLockStrategy(redisRepo), nil

	case "pessimistic":
		log.Println("Using PESSIMISTIC locking strategy")
		return inventory.NewPessimisticStrategy(redisRepo), nil

	case "optimistic":
		log.Println("Using OPTIMISTIC locking strategy (Lua)")
		return inventory.NewOptimisticStrategy(ctx, redisRepo)

	case "queue":
		log.Println("Using QUEUE-based strategy (FIFO)")
		return inventory.NewQueueStrategy(redisRepo), nil

	default:
		log.Println("Using default OPTIMISTIC locking strategy")
		return inventory.NewOptimisticStrategy(ctx, redisRepo)
	}
}

func setupRouter(
	health *handlers.HealthHandler,
	product *handlers.ProductHandler,
	purchase *handlers.PurchaseHandler,
	orders *handlers.OrderHandler,
) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.MetricsMiddleware())

	// Health endpoints (no rate limiting)
	router.GET("/health", health.Health)
	router.GET("/ready", health.Ready)

	// Metrics endpoint
	router.GET("/metrics", middleware.GetMetrics())

	// Product endpoints
	router.POST("/products", product.Create)
	router.GET("/products", product.GetAll)
	router.GET("/products/:id", product.GetByID)
	router.PUT("/products/:id", product.Update)
	router.DELETE("/products/:id", product.Delete)

	// Purchase endpoint
	router.POST("/purchase", purchase.Purchase)

	// Order endpoints
	router.GET("/orders", orders.GetByUser)
	router.GET("/orders/:id", orders.GetByID)

	// Admin endpoints
	router.POST("/admin/sync-inventory", product.SyncInventory)

	return router
}
