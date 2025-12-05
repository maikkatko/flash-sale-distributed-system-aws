package processor

import (
	"context"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// PaymentRequest contains payment details
type PaymentRequest struct {
	OrderID   string
	UserID    string
	Amount    float64
	ProductID int
}

// PaymentResult contains the outcome of payment processing
type PaymentResult struct {
	Success       bool
	TransactionID string
	Reason        string
}

// PaymentProcessor simulates payment processing
type PaymentProcessor struct {
	minDelayMs  int
	maxDelayMs  int
	failureRate float64
}

// NewPaymentProcessor creates a new payment processor with configurable behavior
func NewPaymentProcessor() *PaymentProcessor {
	return &PaymentProcessor{
		minDelayMs:  getEnvInt("PAYMENT_MIN_DELAY_MS", 100),
		maxDelayMs:  getEnvInt("PAYMENT_MAX_DELAY_MS", 500),
		failureRate: getEnvFloat("PAYMENT_FAILURE_RATE", 0.0), // Default: no failures
	}
}

// ProcessPayment simulates a payment with configurable delay and failure rate
func (p *PaymentProcessor) ProcessPayment(ctx context.Context, req PaymentRequest) (*PaymentResult, error) {
	log.Printf("Processing payment for order %s, amount: $%.2f", req.OrderID, req.Amount)

	// Simulate processing delay
	delay := p.minDelayMs + rand.Intn(p.maxDelayMs-p.minDelayMs+1)
	select {
	case <-time.After(time.Duration(delay) * time.Millisecond):
		// Processing completed
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Simulate random failures based on configured rate
	if p.failureRate > 0 && rand.Float64() < p.failureRate {
		log.Printf("Payment failed (simulated) for order %s", req.OrderID)
		return &PaymentResult{
			Success: false,
			Reason:  "Payment declined by processor (simulated failure)",
		}, nil
	}

	// Generate mock transaction ID
	transactionID := generateTransactionID()

	log.Printf("Payment successful for order %s, transaction: %s", req.OrderID, transactionID)

	return &PaymentResult{
		Success:       true,
		TransactionID: transactionID,
	}, nil
}

func generateTransactionID() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 12)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return "TXN-" + string(b)
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultVal
}
