package consumer

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"order-processor/internal/processor"
)

// OrderMessage represents the message structure from SQS
type OrderMessage struct {
	OrderID        string  `json:"order_id"`
	UserID         string  `json:"user_id"`
	ProductID      int     `json:"product_id"`
	Quantity       int     `json:"quantity"`
	TotalPrice     float64 `json:"total_price"`
	IdempotencyKey string  `json:"idempotency_key"`
	CorrelationID  string  `json:"correlation_id"`
	Timestamp      string  `json:"timestamp"`
}

// SQSConsumer handles polling messages from SQS
type SQSConsumer struct {
	client            *sqs.Client
	queueURL          string
	dlqURL            string
	processor         *processor.OrderProcessor
	maxMessages       int32
	visibilityTimeout int32
	waitTimeSeconds   int32
}

// NewSQSConsumer creates a new SQS consumer
func NewSQSConsumer(proc *processor.OrderProcessor) (*SQSConsumer, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}

	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		log.Fatal("SQS_QUEUE_URL environment variable is required")
	}

	dlqURL := os.Getenv("SQS_DLQ_URL")

	return &SQSConsumer{
		client:            sqs.NewFromConfig(cfg),
		queueURL:          queueURL,
		dlqURL:            dlqURL,
		processor:         proc,
		maxMessages:       getEnvInt32("SQS_MAX_MESSAGES", 10),
		visibilityTimeout: getEnvInt32("SQS_VISIBILITY_TIMEOUT", 300), // 5 minutes
		waitTimeSeconds:   getEnvInt32("SQS_WAIT_TIME", 20),           // Long polling
	}, nil
}

// Start begins polling for messages
func (c *SQSConsumer) Start(ctx context.Context) error {
	log.Printf("Starting SQS consumer for queue: %s", c.queueURL)

	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer context cancelled, stopping...")
			return nil
		default:
			if err := c.pollMessages(ctx); err != nil {
				log.Printf("Error polling messages: %v", err)
				// Brief pause before retrying on error
				time.Sleep(time.Second)
			}
		}
	}
}

func (c *SQSConsumer) pollMessages(ctx context.Context) error {
	output, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(c.queueURL),
		MaxNumberOfMessages:   c.maxMessages,
		VisibilityTimeout:     c.visibilityTimeout,
		WaitTimeSeconds:       c.waitTimeSeconds,
		MessageAttributeNames: []string{"All"},
	})
	if err != nil {
		return err
	}

	for _, msg := range output.Messages {
		if err := c.processMessage(ctx, msg); err != nil {
			log.Printf("Failed to process message %s: %v", *msg.MessageId, err)
			// Move to DLQ if configured
			if c.dlqURL != "" {
				c.moveToDLQ(ctx, msg, err.Error())
			}
		}
		// Delete message after processing (success or moved to DLQ)
		c.deleteMessage(ctx, msg)
	}

	return nil
}

func (c *SQSConsumer) processMessage(ctx context.Context, msg types.Message) error {
	if msg.Body == nil {
		return nil
	}

	var orderMsg OrderMessage
	if err := json.Unmarshal([]byte(*msg.Body), &orderMsg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return err
	}

	log.Printf("Processing order: %s (correlation_id: %s)", orderMsg.OrderID, orderMsg.CorrelationID)

	// Convert to processor request
	req := processor.ProcessRequest{
		OrderID:        orderMsg.OrderID,
		UserID:         orderMsg.UserID,
		ProductID:      orderMsg.ProductID,
		Quantity:       orderMsg.Quantity,
		TotalPrice:     orderMsg.TotalPrice,
		IdempotencyKey: orderMsg.IdempotencyKey,
		CorrelationID:  orderMsg.CorrelationID,
	}

	return c.processor.Process(ctx, req)
}

func (c *SQSConsumer) deleteMessage(ctx context.Context, msg types.Message) {
	_, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: msg.ReceiptHandle,
	})
	if err != nil {
		log.Printf("Failed to delete message %s: %v", *msg.MessageId, err)
	}
}

func (c *SQSConsumer) moveToDLQ(ctx context.Context, msg types.Message, errorReason string) {
	if c.dlqURL == "" {
		return
	}

	// Add error info to message attributes
	_, err := c.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(c.dlqURL),
		MessageBody: msg.Body,
		MessageAttributes: map[string]types.MessageAttributeValue{
			"ErrorReason": {
				DataType:    aws.String("String"),
				StringValue: aws.String(errorReason),
			},
			"OriginalMessageId": {
				DataType:    aws.String("String"),
				StringValue: msg.MessageId,
			},
		},
	})
	if err != nil {
		log.Printf("Failed to move message to DLQ: %v", err)
	} else {
		log.Printf("Message %s moved to DLQ", *msg.MessageId)
	}
}

func getEnvInt32(key string, defaultVal int32) int32 {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return int32(i)
		}
	}
	return defaultVal
}
