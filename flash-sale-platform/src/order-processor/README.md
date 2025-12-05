# Order Processor Service

Background worker that consumes orders from SQS, processes payments, and persists to MySQL.

## Environment Variables

### Required
| Variable | Description | Example |
|----------|-------------|---------|
| `SQS_QUEUE_URL` | SQS queue URL for orders | `https://sqs.us-east-1.amazonaws.com/123456789/flash-sale-orders` |
| `DB_HOST` | MySQL host | `flash-sale-db.xxx.us-east-1.rds.amazonaws.com` |
| `DB_USER` | Database username | `admin` |
| `DB_PASSWORD` | Database password | `SecurePassword123!` |
| `DB_NAME` | Database name | `flashsale` |
| `REDIS_ADDR` | Redis address | `flash-sale-redis.xxx.cache.amazonaws.com:6379` |

### Optional
| Variable | Description | Default |
|----------|-------------|---------|
| `AWS_REGION` | AWS region | `us-east-1` |
| `DB_PORT` | Database port | `3306` |
| `REDIS_PASSWORD` | Redis password | `` |
| `SQS_DLQ_URL` | Dead letter queue URL | `` |
| `SQS_MAX_MESSAGES` | Max messages per poll | `10` |
| `SQS_VISIBILITY_TIMEOUT` | Message visibility timeout (seconds) | `300` |
| `SQS_WAIT_TIME` | Long polling wait time (seconds) | `20` |
| `PAYMENT_MIN_DELAY_MS` | Min payment simulation delay | `100` |
| `PAYMENT_MAX_DELAY_MS` | Max payment simulation delay | `500` |
| `PAYMENT_FAILURE_RATE` | Payment failure rate (0.0-1.0) | `0.0` |
| `SHUTDOWN_TIMEOUT` | Graceful shutdown timeout | `30s` |

## Processing Flow

1. **Poll SQS** - Long-poll for order messages
2. **Check Idempotency** - Skip if already processed
3. **Process Payment** - Simulate payment with configurable delay/failure
4. **Persist Order** - Insert order into PostgreSQL
5. **Update Stock** - Decrement product stock in database
6. **Cleanup** - Delete Redis reservation, set idempotency key
7. **Delete Message** - Remove from SQS on success

## Building

```bash
# Local build
go build -o order-processor ./cmd/processor

# Docker build
docker build -t order-processor .
```

## Running Locally

```bash
export SQS_QUEUE_URL="https://sqs.us-east-1.amazonaws.com/123456789/flash-sale-orders"
export DB_HOST="localhost"
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="flashsale"
export REDIS_ADDR="localhost:6379"

./order-processor
```

## Message Format

Expected SQS message body:
```json
{
  "order_id": "uuid",
  "user_id": "user123",
  "product_id": 1,
  "quantity": 1,
  "total_price": 99.99,
  "idempotency_key": "unique-key",
  "correlation_id": "uuid",
  "timestamp": "2024-01-15T10:30:00Z"
}
```