package order

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// OrderMessage is the message published to SQS for async processing
type OrderMessage struct {
	OrderID        string  `json:"order_id"`
	UserID         string  `json:"user_id"`
	ProductID      int     `json:"product_id"`
	Quantity       int     `json:"quantity"`
	TotalPrice     float64 `json:"total_price"`
	IdempotencyKey string  `json:"idempotency_key"`
	Timestamp      string  `json:"timestamp"`
}

// Publisher handles publishing orders to SQS
type Publisher struct {
	client   *sqs.Client
	queueURL string
}

// NewPublisher creates a new SQS publisher
func NewPublisher(client *sqs.Client, queueURL string) *Publisher {
	return &Publisher{
		client:   client,
		queueURL: queueURL,
	}
}

// Publish sends an order message to SQS
func (p *Publisher) Publish(ctx context.Context, msg OrderMessage) error {
	if p.queueURL == "" {
		// SQS not configured - skip (useful for local dev)
		return nil
	}

	if msg.Timestamp == "" {
		msg.Timestamp = time.Now().Format(time.RFC3339)
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	_, err = p.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(p.queueURL),
		MessageBody: aws.String(string(body)),
	})
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

// IsConfigured returns true if SQS is configured
func (p *Publisher) IsConfigured() bool {
	return p.queueURL != ""
}
