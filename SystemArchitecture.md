```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              LOAD TESTING LAYER                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────────┐  │
│  │     Locust      │  │   Bot Simulator │  │    Metrics Collector        │  │
│  │   (500 users)   │  │  (Exp 4: Bots)  │  │  (Custom CloudWatch Agent)  │  │
│  └────────┬────────┘  └────────┬────────┘  └─────────────────────────────┘  │
└───────────┼────────────────────┼────────────────────────────────────────────┘
            │                    │
            ▼                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                               EDGE LAYER                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                    Application Load Balancer                         │   │
│  │  • Health checks (/health)     • Sticky sessions (optional)          │   │
│  │  • TLS termination             • Connection draining                 │   │
│  │  • Request routing             • WAF integration (rate limiting)     │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                    API Gateway (Optional - Exp 4)                    │   │
│  │  • Token bucket rate limiting   • Request throttling                 │   │
│  │  • Bot detection headers        • Queue-based admission control      │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          COMPUTE LAYER (ECS Fargate)                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │                        Flash Sale API Service                      │     │
│  │  ┌─────────────────────────────────────────────────────────────┐   │     │
│  │  │  Endpoints:                                                 │   │     │
│  │  │  • GET  /products          - List available products        │   │     │
│  │  │  • GET  /products/{id}     - Product details + stock        │   │     │
│  │  │  • POST /purchase          - Attempt purchase               │   │     │
│  │  │  • GET  /orders/{id}       - Order status                   │   │     │
│  │  │  • GET  /health            - Health check                   │   │     │
│  │  │  • GET  /metrics           - Prometheus metrics             │   │     │
│  │  └─────────────────────────────────────────────────────────────┘   │     │
│  │  ┌─────────────────────────────────────────────────────────────┐   │     │
│  │  │  Inventory Strategies (Feature Flagged):                    │   │     │
│  │  │  • Strategy A: No Locking (baseline)                        │   │     │
│  │  │  • Strategy B: Pessimistic (Redis SETNX distributed lock)   │   │     │
│  │  │  • Strategy C: Optimistic (Redis WATCH/MULTI or Lua scripts)│   │     │
│  │  │  • Strategy D: Queue-based (SQS serialization)              │   │     │
│  │  └─────────────────────────────────────────────────────────────┘   │     │
│  │  ┌─────────────────────────────────────────────────────────────┐   │     │
│  │  │  Resilience Patterns:                                       │   │     │
│  │  │  • Circuit breaker (Redis, DB connections)                  │   │     │
│  │  │  • Retry with exponential backoff                           │   │     │
│  │  │  • Request timeout (context deadline)                       │   │     │
│  │  │  • Bulkhead (connection pool limits)                        │   │     │
│  │  └─────────────────────────────────────────────────────────────┘   │     │
│  │                                                                    │     │
│  │  Tech: Go + Gin | Min: 2 | Max: 10 tasks | CPU: 512 | Mem: 1024    │     │
│  └────────────────────────────────────────────────────────────────────┘     │
│                           │                                                 │
│           ┌───────────────┼───────────────┐                                 │
│           │               │               │                                 │
│           ▼               ▼               ▼                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────────────┐      │
│  │ Check Stock │  │  Acquire    │  │    Publish to SQS               │      │
│  │   (Redis)   │  │    Lock     │  │    (async order processing)     │      │
│  └─────────────┘  └─────────────┘  └─────────────────────────────────┘      │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
        ┌───────────────────────────┼───────────────────────────┐
        │                           │                           │
        ▼                           ▼                           ▼
┌───────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│   MESSAGE LAYER   │    │    CACHE LAYER      │    │   DATABASE LAYER    │
├───────────────────┤    ├─────────────────────┤    ├─────────────────────┤
│                   │    │                     │    │                     │
│  ┌─────────────┐  │    │  ┌───────────────┐  │    │  ┌───────────────┐  │
│  │ Order Queue │  │    │  │    Redis      │  │    │  │RDS PostgreSQL │  │
│  │   (SQS)     │  │    │  │ (ElastiCache) │  │    │  │   (Aurora)    │  │
│  │             │  │    │  │               │  │    │  │               │  │
│  │ • Standard  │  │    │  │ Key Patterns: │  │    │  │ Tables:       │  │
│  │ • Vis: 5min │  │    │  │               │  │    │  │ • products    │  │
│  │ • MaxRcv: 3 │  │    │  │ inv:{pid}     │  │    │  │ • orders      │  │
│  └──────┬──────┘  │    │  │   → stock cnt │  │    │  │ • order_items │  │
│         │         │    │  │               │  │    │  │ • audit_log   │  │
│         │         │    │  │ lock:{pid}    │  │    │  │               │  │
│  ┌──────▼──────┐  │    │  │   → dist lock │  │    │  │ Indexes:      │  │
│  │     DLQ     │  │    │  │               │  │    │  │ • product_id  │  │
│  │             │  │    │  │ res:{uid}:{p} │  │    │  │ • user_id     │  │
│  │ Failed ords │  │    │  │   → reservatn │  │    │  │ • created_at  │  │
│  │ for analysis│  │    │  │               │  │    │  │               │  │
│  └─────────────┘  │    │  │ idem:{key}    │  │    │  │ Multi-AZ      │  │
│                   │    │  │   → dedup     │  │    │  │ Read Replica  │  │
│  ┌─────────────┐  │    │  │               │  │    │  │   (optional)  │  │
│  │ Fairness Q  │  │    │  │ queue:{pid}   │  │    │  └───────────────┘  │
│  │  (Exp 4)    │  │    │  │   → FIFO list │  │    │                     │
│  │             │  │    │  │               │  │    │                     │
│  │ FIFO queue  │  │    │  │ circuit:{svc} │  │    │                     │
│  │ for ordered │  │    │  │   → breaker   │  │    │                     │
│  │ processing  │  │    │  │     state     │  │    │                     │
│  └─────────────┘  │    │  │               │  │    │                     │
│                   │    │  │ Lua Scripts:  │  │    │                     │
└───────────────────┘    │  │ • atomic_decr │  │    └─────────────────────┘
                         │  │ • check_and_  │  │
                         │  │   reserve     │  │
                         │  │ • release_    │  │
                         │  │   lock        │  │
                         │  └───────────────┘  │
                         │                     │
                         └─────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                        ASYNC PROCESSING LAYER                               │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │                      Order Processor Service                       │     │
│  │                                                                    │     │
│  │  • Poll SQS for pending orders                                     │     │
│  │  • Validate reservation still valid                                │     │
│  │  • Simulate payment processing (configurable delay)                │     │
│  │  • Persist to RDS (transactional)                                  │     │
│  │  • Cleanup Redis reservation                                       │     │
│  │  • Update inventory in DB (final sync)                             │     │
│  │  • Emit completion metrics                                         │     │
│  │                                                                    │     │
│  │  Tech: Go | Min: 1 | Max: 5 tasks | CPU: 256 | Mem: 512            │     │
│  └────────────────────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                         OBSERVABILITY LAYER                                  │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌───────────────────┐  ┌──────────────────┐  ┌──────────────────────────┐   │
│  │    CloudWatch     │  │   CloudWatch     │  │    Custom Metrics        │   │
│  │ Container Insights│  │      Logs        │  │                          │   │
│  │                   │  │                  │  │  • Oversell counter      │   │
│  │ • CPU/Memory      │  │ • Application    │  │  • Purchase success rate │   │
│  │ • Network I/O     │  │   logs (JSON)    │  │  • Latency percentiles   │   │
│  │ • Task count      │  │ • Access logs    │  │  • Queue depth           │   │
│  │ • Request count   │  │ • Error traces   │  │  • Lock contention       │   │
│  └───────────────────┘  └──────────────────┘  │  • Fairness metrics      │   │
│                                               │  • Circuit breaker state │   │
│  ┌──────────────────┐  ┌──────────────────┐   └──────────────────────────┘   │
│  │  CloudWatch      │  │    X-Ray         │                                  │
│  │    Alarms        │  │   (Tracing)      │  ┌──────────────────────────┐    │
│  │                  │  │                  │  │   Experiment Metrics     │    │
│  │ • High error %   │  │ • Request flow   │  │                          │    │
│  │ • Scaling events │  │ • Latency trace  │  │                          │    │
│  │ • DLQ depth      │  │ • Service map    │  │  • Raw Locust CSV        │    │
│  │ • Redis latency  │  │ • Error analysis │  │  • CloudWatch exports    │    │
│  └──────────────────┘  └──────────────────┘  │  • Trial comparisons     │    │
│                                              └──────────────────────────┘    │
└──────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                       CHAOS ENGINEERING LAYER (Exp 3)                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │                     Fault Injection Mechanisms                     │     │
│  │                                                                    │     │
│  │  Application-Level (Feature Flags):                                │     │
│  │  • CHAOS_REDIS_FAIL=true      → Simulate Redis connection failure  │     │
│  │  • CHAOS_DB_LATENCY=500ms     → Add artificial DB latency          │     │
│  │  • CHAOS_RANDOM_PANIC=0.01    → 1% chance of service panic         │     │
│  │  • CHAOS_PAYMENT_FAIL=0.1     → 10% payment failure rate           │     │
│  │                                                                    │     │
│  │  Infrastructure-Level:                                             │     │
│  │  • ECS task termination (via AWS CLI/SDK)                          │     │
│  │  • Redis failover trigger                                          │     │
│  │  • RDS failover (Multi-AZ)                                         │     │
│  │  • Security group modification (network partition simulation)      │     │
│  │                                                                    │     │
│  └────────────────────────────────────────────────────────────────────┘     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

## Auto-Scaling Configuration (Exp 2)

┌─────────────────────────────────────────────────────────────────────────────┐
│                        SCALING POLICIES                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Policy A: Target Tracking (CPU-based)                                      │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │  target_value          = 60  (60% CPU utilization)                 │     │
│  │  scale_in_cooldown     = 300 seconds                               │     │
│  │  scale_out_cooldown    = 60  seconds                               │     │
│  │  disable_scale_in      = false                                     │     │
│  └────────────────────────────────────────────────────────────────────┘     │
│                                                                             │
│  Policy B: Step Scaling (Request-based)                                     │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │  Alarm: RequestCountPerTarget > 1000                               │     │
│  │  Steps:                                                            │     │
│  │    • 1000-2000 req → +2 tasks                                      │     │
│  │    • 2000-3000 req → +3 tasks                                      │     │
│  │    • 3000+     req → +5 tasks (to max)                             │     │
│  │  Cooldown: 30 seconds                                              │     │
│  └────────────────────────────────────────────────────────────────────┘     │
│                                                                             │
│  Policy C: Scheduled + Predictive (Pre-warming)                             │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │  • T-5min before sale: Scale to max (10 tasks)                     │     │
│  │  • T+0 (sale start): Maintain max                                  │     │
│  │  • T+30min: Allow scale-in based on actual load                    │     │
│  └────────────────────────────────────────────────────────────────────┘     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Data Flow Summary

1. User Request → ALB → Flash Sale API
2. API checks Redis for stock (inv:{pid})
3. Based on strategy:
   - No Lock: Direct decrement (race condition possible)
   - Pessimistic: Acquire lock:{pid}, check, decrement, release
   - Optimistic: Lua script atomic check-and-decrement
   - Queue-based: Enqueue to fairness queue, process serially
4. Create reservation in Redis (res:{uid}:{pid}, TTL: 5min)
5. Publish order to SQS
6. Order Processor consumes, validates, processes payment
7. On success: Persist to RDS, clear reservation
8. On failure: Release reservation, increment stock, send to DLQ

## Key Additions

1. **Chaos Engineering Layer** - Missing from original
2. **Fairness Queue (SQS FIFO)** - For Experiment 4
3. **Multiple Scaling Policies** - For Experiment 2 comparison
4. **Circuit Breaker State in Redis** - For resilience tracking
5. **X-Ray Tracing** - For latency analysis
6. **Bot Simulator** - For fairness experiments
7. **S3 Metrics Export** - For offline analysis
8. **Audit Log Table** - For consistency verification
