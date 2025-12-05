# Flash Sale Platform - Implementation Plan

## Phase 1: Foundation Infrastructure

**Goal:** Deploy base infrastructure with Terraform, establish CI/CD

### Step 1.1: Terraform Foundation

```
terraform/
├── modules/
│   ├── vpc/           # VPC, subnets, NAT gateway
│   ├── ecs/           # Cluster, task definitions, services
│   ├── alb/           # Load balancer, target groups, listeners
│   ├── rds/           # Aurora PostgreSQL, parameter groups
│   ├── elasticache/   # Redis cluster, subnet group
│   ├── sqs/           # Order queue, DLQ, fairness queue
│   └── monitoring/    # CloudWatch dashboards, alarms
└── variables.tf
```

**Deliverables:**

- [ ] VPC with public/private subnets across 2 AZs
- [ ] ECS Fargate cluster with service discovery
- [ ] ALB with health check configuration
- [ ] RDS Aurora PostgreSQL (Multi-AZ, db.t3.medium)
- [ ] ElastiCache Redis (cache.t3.micro, cluster mode disabled)
- [ ] SQS queues (standard + DLQ)
- [ ] IAM roles with least-privilege policies
- [ ] ECR repositories for both services

### Step 1.2: Database Schema

```sql
-- products table
CREATE TABLE products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name)
);

-- orders table
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    product_id INT REFERENCES products(id),
    quantity INT NOT NULL DEFAULT 1,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    idempotency_key VARCHAR(255) UNIQUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- audit_log for consistency verification
CREATE TABLE audit_log (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    product_id INT,
    order_id UUID,
    stock_before INT,
    stock_after INT,
    details JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- indexes
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_product_id ON orders(product_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_audit_product_id ON audit_log(product_id);
```

---

## Phase 2: Core Services Development

**Goal:** Implement Flash Sale API and Order Processor with basic functionality

### Step 2.1: Flash Sale API Service

```
flash-sale-api/
├── cmd/
│   └── server/main.go
├── internal/
│   ├── handlers/
│   │   ├── products.go      # GET /products, GET /products/{id}
│   │   ├── purchase.go      # POST /purchase
│   │   ├── orders.go        # GET /orders/{id}
│   │   └── health.go        # GET /health, /ready
│   ├── services/
│   │   ├── inventory/
│   │   │   ├── strategy.go  # Interface definition
│   │   │   ├── none.go      # No locking (baseline)
│   │   │   ├── pessimistic.go
│   │   │   ├── optimistic.go
│   │   │   └── queue.go
│   │   └── order/
│   │       └── publisher.go # SQS publisher
│   ├── repository/
│   │   ├── product.go
│   │   └── redis.go
│   ├── middleware/
│   │   ├── logging.go
│   │   ├── metrics.go
│   │   └── ratelimit.go
│   └── config/
│       └── config.go
├── pkg/
│   ├── circuit/             # Circuit breaker implementation
│   └── retry/               # Retry with backoff
├── Dockerfile
└── go.mod
```

**Key Implementation: Inventory Strategy Interface**

```go
type InventoryStrategy interface {
    // AttemptPurchase tries to reserve inventory
    // Returns (success, error)
    AttemptPurchase(ctx context.Context, req PurchaseRequest) (*Reservation, error)

    // ReleasePurchase releases a failed reservation
    ReleasePurchase(ctx context.Context, reservation *Reservation) error

    // Name returns strategy identifier for metrics
    Name() string
}

type PurchaseRequest struct {
    UserID        string
    ProductID     int
    Quantity      int
    IdempotencyKey string
}

type Reservation struct {
    ID         string
    UserID     string
    ProductID  int
    Quantity   int
    ExpiresAt  time.Time
}
```

### Step 2.2: Order Processor Service

```
order-processor/
├── cmd/
│   └── processor/main.go
├── internal/
│   ├── consumer/
│   │   └── sqs.go          # SQS polling with graceful shutdown
│   ├── processor/
│   │   ├── order.go        # Order processing logic
│   │   └── payment.go      # Payment simulation
│   └── repository/
│       ├── order.go
│       └── redis.go
├── Dockerfile
└── go.mod
```

**Deliverables:**

- [ ] Flash Sale API with all endpoints
- [ ] Connection pooling for PostgreSQL and Redis
- [ ] Request/response logging middleware
- [ ] Prometheus metrics endpoint
- [ ] Order Processor with SQS consumer
- [ ] Simulated payment processing (configurable delay)
- [ ] Unit tests for critical paths

---

## Phase 3: Inventory Strategies

**Goal:** Implement all four inventory management strategies for Experiment 1

### Step 3.1: Strategy A - No Locking (Baseline)

```go
func (s *NoLockStrategy) AttemptPurchase(ctx context.Context, req PurchaseRequest) (*Reservation, error) {
    // Direct Redis DECR - will cause overselling under contention
    stock, err := s.redis.Decr(ctx, fmt.Sprintf("inv:%d", req.ProductID)).Result()
    if err != nil {
        return nil, err
    }
    if stock < 0 {
        // Rollback
        s.redis.Incr(ctx, fmt.Sprintf("inv:%d", req.ProductID))
        return nil, ErrOutOfStock
    }
    return &Reservation{...}, nil
}
```

### Step 3.2: Strategy B - Pessimistic Locking

```go
func (s *PessimisticStrategy) AttemptPurchase(ctx context.Context, req PurchaseRequest) (*Reservation, error) {
    lockKey := fmt.Sprintf("lock:%d", req.ProductID)
    lockValue := uuid.NewString()

    // Acquire distributed lock with SETNX
    acquired, err := s.redis.SetNX(ctx, lockKey, lockValue, 5*time.Second).Result()
    if err != nil || !acquired {
        return nil, ErrLockNotAcquired
    }
    defer s.releaseLock(ctx, lockKey, lockValue)

    // Check and decrement under lock
    stock, _ := s.redis.Get(ctx, fmt.Sprintf("inv:%d", req.ProductID)).Int()
    if stock < req.Quantity {
        return nil, ErrOutOfStock
    }
    s.redis.DecrBy(ctx, fmt.Sprintf("inv:%d", req.ProductID), int64(req.Quantity))

    return &Reservation{...}, nil
}
```

### Step 3.3: Strategy C - Optimistic Locking (Lua Script)

```lua
-- atomic_purchase.lua
local stock_key = KEYS[1]
local quantity = tonumber(ARGV[1])

local current = tonumber(redis.call('GET', stock_key) or 0)
if current < quantity then
    return {0, current}  -- failure, return current stock
end

local new_stock = redis.call('DECRBY', stock_key, quantity)
return {1, new_stock}  -- success, return new stock
```

```go
func (s *OptimisticStrategy) AttemptPurchase(ctx context.Context, req PurchaseRequest) (*Reservation, error) {
    result, err := s.redis.EvalSha(ctx, s.luaScriptSHA,
        []string{fmt.Sprintf("inv:%d", req.ProductID)},
        req.Quantity,
    ).Result()

    res := result.([]interface{})
    if res[0].(int64) == 0 {
        return nil, ErrOutOfStock
    }
    return &Reservation{...}, nil
}
```

### Step 3.4: Strategy D - Queue-Based

```go
func (s *QueueStrategy) AttemptPurchase(ctx context.Context, req PurchaseRequest) (*Reservation, error) {
    // Add to per-product FIFO queue in Redis
    queueKey := fmt.Sprintf("queue:%d", req.ProductID)
    position, _ := s.redis.RPush(ctx, queueKey, req.UserID).Result()

    // Wait for turn with timeout
    ticker := time.NewTicker(50 * time.Millisecond)
    timeout := time.After(10 * time.Second)

    for {
        select {
        case <-timeout:
            s.redis.LRem(ctx, queueKey, 1, req.UserID)
            return nil, ErrTimeout
        case <-ticker.C:
            front, _ := s.redis.LIndex(ctx, queueKey, 0).Result()
            if front == req.UserID {
                // Our turn - process
                defer s.redis.LPop(ctx, queueKey)
                return s.processInQueue(ctx, req)
            }
        }
    }
}
```

**Deliverables:**

- [ ] All four strategies implemented behind interface
- [ ] Feature flag to switch strategies (env var: INVENTORY_STRATEGY)
- [ ] Strategy-specific metrics (lock wait time, queue depth, etc.)
- [ ] Integration tests for each strategy

---

## Phase 4: Resilience Patterns

**Goal:** Implement circuit breakers, retries, and chaos engineering for Experiment 3

### Step 4.1: Circuit Breaker Implementation

```go
type CircuitBreaker struct {
    name          string
    maxFailures   int
    resetTimeout  time.Duration
    state         State  // Closed, Open, HalfOpen
    failures      int
    lastFailure   time.Time
    mu            sync.RWMutex
    redis         *redis.Client  // For distributed state
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    if !cb.AllowRequest() {
        return ErrCircuitOpen
    }

    err := fn()
    if err != nil {
        cb.RecordFailure()
        return err
    }
    cb.RecordSuccess()
    return nil
}
```

### Step 4.2: Chaos Engineering Hooks

```go
type ChaosConfig struct {
    RedisFailRate    float64 `env:"CHAOS_REDIS_FAIL" default:"0"`
    DBLatencyMs      int     `env:"CHAOS_DB_LATENCY" default:"0"`
    RandomPanicRate  float64 `env:"CHAOS_RANDOM_PANIC" default:"0"`
    PaymentFailRate  float64 `env:"CHAOS_PAYMENT_FAIL" default:"0"`
}

func (c *ChaosMiddleware) MaybeInjectFault(service string) error {
    switch service {
    case "redis":
        if rand.Float64() < c.config.RedisFailRate {
            return ErrChaosInjected
        }
    case "db":
        if c.config.DBLatencyMs > 0 {
            time.Sleep(time.Duration(c.config.DBLatencyMs) * time.Millisecond)
        }
    }
    return nil
}
```

### Step 4.3: Chaos Test Scripts

```bash
# chaos/redis_failure.sh
#!/bin/bash
# Simulate Redis connection failure by modifying security group

REDIS_SG_ID=$(terraform output -raw redis_security_group_id)
API_SG_ID=$(terraform output -raw api_security_group_id)

echo "Blocking Redis access..."
aws ec2 revoke-security-group-ingress \
    --group-id $REDIS_SG_ID \
    --source-group $API_SG_ID \
    --protocol tcp --port 6379

sleep 60  # Observe behavior

echo "Restoring Redis access..."
aws ec2 authorize-security-group-ingress \
    --group-id $REDIS_SG_ID \
    --source-group $API_SG_ID \
    --protocol tcp --port 6379
```

**Deliverables:**

- [ ] Circuit breaker for Redis and DB connections
- [ ] Retry with exponential backoff
- [ ] Chaos injection via environment variables
- [ ] Infrastructure chaos scripts (Redis, RDS failover, task kill)
- [ ] MTTR measurement instrumentation

---

## Phase 5: Auto-Scaling Configuration

**Goal:** Configure and test multiple scaling policies for Experiment 2

### Step 5.1: Terraform Scaling Policies

```hcl
# Policy A: Target Tracking (CPU)
resource "aws_appautoscaling_policy" "cpu_target_tracking" {
  name               = "flash-sale-cpu-tracking"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.api.resource_id
  scalable_dimension = aws_appautoscaling_target.api.scalable_dimension
  service_namespace  = aws_appautoscaling_target.api.service_namespace

  target_tracking_scaling_policy_configuration {
    target_value       = 60
    scale_in_cooldown  = 300
    scale_out_cooldown = 60

    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
  }
}

# Policy B: Step Scaling (Request-based)
resource "aws_appautoscaling_policy" "request_step" {
  name               = "flash-sale-request-step"
  policy_type        = "StepScaling"
  resource_id        = aws_appautoscaling_target.api.resource_id
  scalable_dimension = aws_appautoscaling_target.api.scalable_dimension
  service_namespace  = aws_appautoscaling_target.api.service_namespace

  step_scaling_policy_configuration {
    adjustment_type         = "ChangeInCapacity"
    cooldown                = 30
    metric_aggregation_type = "Average"

    step_adjustment {
      metric_interval_lower_bound = 0
      metric_interval_upper_bound = 1000
      scaling_adjustment          = 2
    }
    step_adjustment {
      metric_interval_lower_bound = 1000
      metric_interval_upper_bound = 2000
      scaling_adjustment          = 3
    }
    step_adjustment {
      metric_interval_lower_bound = 2000
      scaling_adjustment          = 5
    }
  }
}

# CloudWatch Alarm for Step Scaling
resource "aws_cloudwatch_metric_alarm" "high_request_count" {
  alarm_name          = "flash-sale-high-requests"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "RequestCountPerTarget"
  namespace           = "AWS/ApplicationELB"
  period              = 60
  statistic           = "Sum"
  threshold           = 1000
  alarm_actions       = [aws_appautoscaling_policy.request_step.arn]
}
```

### Step 5.2: Scaling Test Script

```python
# test_scaling.py
import subprocess
import time
import boto3

def measure_scale_time(policy_name: str, target_users: int):
    """Measure time from traffic spike to scaled state"""
    ecs = boto3.client('ecs')
    cloudwatch = boto3.client('cloudwatch')

    # Record initial task count
    initial_count = get_task_count(ecs)

    # Start Locust spike
    start_time = time.time()
    locust_proc = subprocess.Popen([
        'locust', '-f', 'locustfile.py',
        '--headless', '-u', str(target_users),
        '-r', str(target_users),  # All users spawn immediately
        '-t', '5m'
    ])

    # Monitor scaling
    scale_out_time = None
    target_reached_time = None

    while time.time() - start_time < 300:  # 5 min max
        current = get_task_count(ecs)

        if current > initial_count and scale_out_time is None:
            scale_out_time = time.time() - start_time

        if current >= 10:  # Max tasks
            target_reached_time = time.time() - start_time
            break

        time.sleep(5)

    locust_proc.terminate()

    return {
        'policy': policy_name,
        'initial_tasks': initial_count,
        'time_to_first_scale': scale_out_time,
        'time_to_max': target_reached_time,
        'error_rate_during_scale': get_error_rate(cloudwatch, start_time)
    }
```

**Deliverables:**

- [ ] Target tracking policy (CPU-based)
- [ ] Step scaling policy (request-based)
- [ ] Scheduled scaling (pre-warming)
- [ ] Scale test automation script
- [ ] Time-to-scale measurement instrumentation

---

## Phase 6: Load Testing & Metrics

**Goal:** Complete Locust setup and metrics collection

### Step 6.1: Locust Configuration

Normal user vs Aggressive Buyer (Bot)

### Step 6.2: Custom Metrics Dashboard

```python
# metrics/dashboard.tf
resource "aws_cloudwatch_dashboard" "flash_sale" {
  dashboard_name = "FlashSale-Experiments"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        properties = {
          title  = "Purchase Success vs Failure"
          metrics = [
            ["FlashSale", "PurchaseSuccess", "Strategy", "${var.strategy}"],
            ["FlashSale", "PurchaseFailure", "Strategy", "${var.strategy}"],
            ["FlashSale", "Oversell", "Strategy", "${var.strategy}"]
          ]
        }
      },
      {
        type   = "metric"
        properties = {
          title  = "Latency Percentiles"
          metrics = [
            ["FlashSale", "PurchaseLatency", "Percentile", "p50"],
            ["FlashSale", "PurchaseLatency", "Percentile", "p95"],
            ["FlashSale", "PurchaseLatency", "Percentile", "p99"]
          ]
        }
      },
      {
        type   = "metric"
        properties = {
          title  = "ECS Task Count"
          metrics = [
            ["AWS/ECS", "RunningTaskCount", "ServiceName", "flash-sale-api"]
          ]
        }
      }
    ]
  })
}
```

**Deliverables:**

- [ ] Locust file with browsing/purchasing behavior
- [ ] Bot simulator for fairness testing
- [ ] CloudWatch custom metrics emission
- [ ] Dashboard for real-time monitoring
- [ ] S3 export for offline analysis

---

## Phase 7: Experiment Execution

**Goal:** Run all experiments and collect data

### Experiment 1: Locking Strategy Comparison

```bash
# run_experiment1.sh
STRATEGIES=("none" "pessimistic" "optimistic" "queue")
CONTENTION_RATIOS=(10 20 50)  # users per item
ITEMS=10
TRIALS=5

for strategy in "${STRATEGIES[@]}"; do
  for ratio in "${CONTENTION_RATIOS[@]}"; do
    users=$((ITEMS * ratio))

    for trial in $(seq 1 $TRIALS); do
      # Reset inventory
      ./scripts/reset_inventory.sh $ITEMS 1  # 1 item each

      # Update strategy
      aws ecs update-service \
        --cluster flash-sale \
        --service api \
        --force-new-deployment \
        --environment "INVENTORY_STRATEGY=$strategy"

      # Wait for deployment
      sleep 60

      # Run load test
      locust -f locustfile.py \
        --headless \
        -u $users -r $users \
        -t 2m \
        --csv "results/exp1/${strategy}_${ratio}_${trial}"

      # Export metrics
      ./scripts/export_metrics.sh "exp1/${strategy}_${ratio}_${trial}"
    done
  done
done
```

### Experiment 2: Auto-Scaling Comparison

```bash
# run_experiment2.sh
POLICIES=("cpu_target" "step_scaling" "scheduled")

for policy in "${POLICIES[@]}"; do
  # Apply policy via Terraform
  terraform apply -var="scaling_policy=$policy" -auto-approve

  # Wait for policy to activate
  sleep 120

  # Reset to minimum tasks
  aws ecs update-service --desired-count 2 ...
  sleep 60

  # Spike traffic 0 → 500
  locust -f locustfile.py \
    --headless \
    -u 500 -r 500 \
    -t 5m \
    --csv "results/exp2/${policy}"

  ./scripts/export_metrics.sh "exp2/${policy}"
done
```

### Experiment 3: Failure Recovery

```bash
# run_experiment3.sh
SCENARIOS=("redis_failure" "db_failover" "task_kill" "network_partition")
CIRCUIT_BREAKER=("enabled" "disabled")

for scenario in "${SCENARIOS[@]}"; do
  for cb in "${CIRCUIT_BREAKER[@]}"; do
    # Configure circuit breaker
    aws ecs update-service \
      --environment "CIRCUIT_BREAKER_ENABLED=$cb"

    # Start baseline load
    locust -f locustfile.py --headless -u 100 -t 10m &
    LOCUST_PID=$!

    sleep 120  # Stabilize

    # Inject fault
    ./chaos/${scenario}.sh &
    FAULT_START=$(date +%s)

    # Wait for recovery
    sleep 120

    kill $LOCUST_PID

    ./scripts/export_metrics.sh "exp3/${scenario}_${cb}"
  done
done
```

### Experiment 4: Fairness Analysis (Optional)

```python
# analyze_fairness.py
import numpy as np
from scipy import stats

def calculate_gini(purchases_per_user: list) -> float:
    """Calculate Gini coefficient (0 = perfect equality, 1 = perfect inequality)"""
    sorted_purchases = sorted(purchases_per_user)
    n = len(sorted_purchases)
    cumulative = np.cumsum(sorted_purchases)
    return (n + 1 - 2 * np.sum(cumulative) / cumulative[-1]) / n

def analyze_fairness(results_path: str):
    df = pd.read_csv(results_path)

    # Group by user type (human vs bot)
    human_purchases = df[df['user_type'] == 'human']['purchases'].tolist()
    bot_purchases = df[df['user_type'] == 'bot']['purchases'].tolist()

    return {
        'overall_gini': calculate_gini(df['purchases'].tolist()),
        'human_success_rate': np.mean(human_purchases) / np.mean(bot_purchases),
        'human_gini': calculate_gini(human_purchases),
        'bot_gini': calculate_gini(bot_purchases)
    }
```

---

## Phase 8: Analysis & Documentation

Final project report

---

## Risk Mitigation

| Risk                         | Mitigation                                       |
| ---------------------------- | ------------------------------------------------ |
| AWS costs                    | Shut down services when not testing              |
| Time constraints             | Exp 4 (fairness) is optional, prioritize Exp 1-3 |
| Debugging distributed issues | X-Ray tracing from day 1                         |
