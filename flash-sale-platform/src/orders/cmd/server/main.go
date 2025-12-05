package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"orders/internal/config"
	"orders/internal/handlers"
	"orders/internal/repository"
	"orders/internal/services/inventory"
)

func main() {
	log.Println("Starting Orders Service...")

	// Load configuration
	cfg := config.Load()

	// Initialize repositories
	productRepo, err := repository.NewProductRepository(
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBName,
	)
	if err != nil {
		log.Fatalf("Failed to initialize product repository: %v", err)
	}
	defer productRepo.Close()

	redisRepo, err := repository.NewRedisRepository(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		log.Fatalf("Failed to initialize Redis repository: %v", err)
	}
	defer redisRepo.Close()

	sqsRepo, err := repository.NewSQSRepository(cfg.AWSRegion, cfg.SQSQueueURL)
	if err != nil {
		log.Fatalf("Failed to initialize SQS repository: %v", err)
	}

	// Create inventory strategy based on configuration
	strategy := inventory.NewStrategy(cfg.LockingStrategy)
	log.Printf("Using locking strategy: %s", strategy.GetName())

	// Initialize handlers
	purchaseHandler := handlers.NewPurchaseHandler(
		productRepo,
		redisRepo,
		sqsRepo,
		strategy,
	)

	// Setup router
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":   "healthy",
			"service":  "orders",
			"strategy": strategy.GetName(),
		})
	})

	// Purchase endpoint
	router.POST("/purchase", purchaseHandler.Purchase)

	// Start server
	log.Printf("Orders Service listening on port %s", cfg.Port)
	log.Printf("Locking strategy: %s", cfg.LockingStrategy)
	router.Run(":" + cfg.Port)
}
