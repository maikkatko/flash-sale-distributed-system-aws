package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"orders/pkg/models"
)

type SQSRepository struct {
	client   *sqs.Client
	queueURL string
	ctx      context.Context
}

func NewSQSRepository(region, queueURL string) (*SQSRepository, error) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	client := sqs.NewFromConfig(cfg)

	log.Println("Connected to SQS")

	return &SQSRepository{
		client:   client,
		queueURL: queueURL,
		ctx:      ctx,
	}, nil
}

// PublishOrder publishes an order message to SQS
func (r *SQSRepository) PublishOrder(msg models.OrderMessage) error {
	messageBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	_, err = r.client.SendMessage(r.ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(r.queueURL),
		MessageBody: aws.String(string(messageBody)),
	})
	if err != nil {
		return fmt.Errorf("failed to send to SQS: %v", err)
	}

	log.Printf("Published order to SQS: User %s, Product %d", msg.UserID, msg.ProductID)
	return nil
}

// PublishOrderWithRetry attempts to publish with exponential backoff
func (r *SQSRepository) PublishOrderWithRetry(msg models.OrderMessage, maxRetries int) error {
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := r.PublishOrder(msg)
		if err == nil {
			return nil
		}

		if attempt < maxRetries-1 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			log.Printf("SQS publish failed (attempt %d/%d), retrying in %v: %v",
				attempt+1, maxRetries, backoff, err)
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("failed after %d retries", maxRetries)
}
