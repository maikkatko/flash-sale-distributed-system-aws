package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order-processor/internal/consumer"
	"order-processor/internal/processor"
	"order-processor/internal/repository"
)

func main() {
	log.Println("Starting Order Processor Service...")

	// Initialize repositories
	orderRepo, err := repository.NewOrderRepository()
	if err != nil {
		log.Fatalf("Failed to initialize order repository: %v", err)
	}
	defer orderRepo.Close()

	redisRepo, err := repository.NewRedisRepository()
	if err != nil {
		log.Fatalf("Failed to initialize redis repository: %v", err)
	}
	defer redisRepo.Close()

	// Initialize payment processor
	paymentProcessor := processor.NewPaymentProcessor()

	// Initialize order processor
	orderProcessor := processor.NewOrderProcessor(orderRepo, redisRepo, paymentProcessor)

	// Initialize SQS consumer
	sqsConsumer, err := consumer.NewSQSConsumer(orderProcessor)
	if err != nil {
		log.Fatalf("Failed to initialize SQS consumer: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start consuming messages
	go func() {
		if err := sqsConsumer.Start(ctx); err != nil {
			log.Printf("Consumer error: %v", err)
			cancel()
		}
	}()

	log.Println("Order Processor is running. Press Ctrl+C to stop.")

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received, stopping gracefully...")

	// Cancel context to stop consumer
	cancel()

	// Give time for in-flight messages to complete
	shutdownTimeout := getEnvDuration("SHUTDOWN_TIMEOUT", 30*time.Second)
	time.Sleep(shutdownTimeout)

	log.Println("Order Processor stopped.")
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
