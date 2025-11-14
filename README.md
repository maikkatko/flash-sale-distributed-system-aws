# C6650-Final-Project

Flash Sale System - High-Contention E-Commerce Platform

## How to Deploy

Navigate to terraform directory and create your own testing.tfvars file as such:

```
db_name     = "flash_sale_db"
db_username = "admin"
db_password = "password"
```

Afterwards, use

```
aws configure
aws configure set aws_session_token $SESSION_TOKEN
```

to authenticate with aws. Afterwards run

```
terraform init
terraform apply -var-file="testing.tfvars"
```

This will deploy the flash sale platform.

With the outputted DNS, run the following to create a product:

```
curl -X POST YOUR_ALB_URL/products \
-H "Content-Type: application/json" \
-d '{
    "name": "Super Widget",
    "description": "A very high-quality widget.",
    "price": 99.99,
    "stock": 100
}'
```

Then GET

```
curl -v YOUR_ALB_URL/products/1
```

Sample output:
$ curl -v curl -v -X POST http://flash-224422.us-west-2.elb.amazonaws.com/products \

> -H "Content-Type: application/json" \
> -d '{
> "name": "Super Widget",
> "description": "A very high-quality widget.",
> "price": 99.99,
> "stock": 100

$ curl -v http://flash-sale-alb-754224422.us-west-2.elb.amazonaws.com/products/1

- Host flash-sale-alb-754224422.us-west-2.elb.amazonaws.com:80 was resolved.
- IPv6: (none)
- IPv4: 54.70.17.29
- Trying 54.70.17.29:80...
- Connected to flash-sale-alb-754224422.us-west-2.elb.amazonaws.com (54.70.17.29) port 80
  > GET /products/1 HTTP/1.1
  > Host: flash-sale-alb-754224422.us-west-2.elb.amazonaws.com
  > User-Agent: curl/8.6.0
  > Accept: _/_
  >
  > < HTTP/1.1 200 OK
  > < Date: Sat, 08 Nov 2025 22:50:48 GMT
  > < Content-Type: application/json; charset=utf-8
  > < Content-Length: 172
  > < Connection: keep-alive
  > <
- Connection #0 to host flash-sale-alb-754224422.us-west-2.elb.amazonaws.com left intact
  {"id":1,"name":"Super Widget","description":"A very high-quality widget.","price":99.99,"stock":100,"created_at":"2025-11-08T22:50:16Z","updated_at":"2025-11-08T22:50:16Z"}

  Topics to explore:

  - Max connections on DB
  - Redis forwarding invalid requests to DB - vulnerability
  - Message queue failure
    - message retries? (blocking, duplication)
    - Dead Letter Queue?


# Flash Sale Testing Flow

## Phase 1: Infrastructure Setup

### Initial Setup: `make setup`

**1. Create Results Directories**
```cmd
mkdir results/
  ├─ formatted_locust_data/
  ├─ cloudwatch_metrics/
  ├─ raw_locust_data/
  ├─ locust_reports/
  └─ charts/
```

**2. Install Python Dependencies**
```bash
pip install -r requirements.txt
  ├─ locust>=2.17.0
  ├─ pandas>=2.0.0
  ├─ matplotlib>=3.7.0
  ├─ boto3>=1.26.0
  ├─ pyyaml>=6.0
  └─ python-dotenv>=1.0.0
```

**3. Initialize Terraform**
```bash
terraform -chdir=flash-sale-platform/terraform init
```

**4. Deploy AWS Infrastructure**
```bash
terraform -chdir=flash-sale-platform/terraform plan
terraform -chdir=flash-sale-platform/terraform apply -auto-approve
  ├─ Creates ECS Cluster
  ├─ Creates Application Load Balancer (ALB)
  ├─ Creates RDS Database
  ├─ Creates Redis Cache
  ├─ Creates RabbitMQ
  ├─ Creates ECS Service with autoscaling policies
  └─ Outputs: ALB_DNS_NAME, ECS_CLUSTER_NAME, etc.
```

**5. Initialize AWS Environment Variables**
```bash
python scripts/init_aws_env_vars.py
  ├─ Queries Terraform outputs
  ├─ Queries AWS resources (ALB DNS, ECS info)
  └─ Writes to .env file:
      - ALB_DNS_NAME
      - ECS_CLUSTER_NAME
      - ECS_SERVICE_NAME
      - SERVICE_NAME
      - etc.
```

---

## Appendix: Test Scenarios Reference

### Baseline Tests
```bash
make test-baseline-target   # 50 users, NormalUser, target_tracking, 5min
make test-baseline-step     # 50 users, NormalUser, step_scaling, 5min
```

### Individual Scenario Tests

**High Contention**
```bash
make test-high-contention   # 100 users, AggressiveBuyer, 60s
```
100 concurrent users competing for products 1-3

**Thundering Herd (Autoscaling)**
```bash
make test-thundering-herd-target  # 500 users, AggressiveBuyer, target_tracking, 3min
make test-thundering-herd-step    # 500 users, AggressiveBuyer, step_scaling, 3min
```
Instant spike from 0→500 users to test autoscaling response

**Sustained Load**
```bash
make test-sustained               # 200 users, NormalUser, 10min
```
Verify stability after scaling up

**Chaos Testing**
```bash
make test-chaos             # 100 users, ChaosTestUser, 5min
```
Moderate load during manual service failure injection

**Fairness Evaluation**
```bash
make test-fairness          # 300 users, AggressiveBuyer, 2min
```
300 users competing for 30 items (10x demand)

---

## Complete Workflow Example

```bash
# 1. Initial Setup (one-time)
make setup

# 2. Seed test data
make seed-data

# 3. Run baseline with target tracking
make test-baseline-target

# 4. Run thundering herd with target tracking
make test-thundering-herd-target

# 5. Switch to step scaling policy
make update-scaling-policy

# 6. Run baseline with step scaling
make test-baseline-step

# 7. Run thundering herd with step scaling
make test-thundering-herd-step

# 8. Analyze results
make analyze

# 9. Cleanup everything
make clean
```

---

## Helper Commands

### View Available Commands
```bash
make help
```

### Terraform Operations
```bash
make init-tf           # Initialize Terraform only
make init-aws-vars     # Refresh .env with AWS resource info
```

---

### Alternative: `make setup-aws`
```bash
python scripts/setup_aws.py       # Deploy AWS infra only
python scripts/init_aws_env_vars.py  # Initialize .env
```

---

### Seed Test Data: `make seed-data`
```bash
python scripts/seed_products.py
  └─ Creates initial product inventory in database
     (Products 1-50 for testing)
```

---

### Update Scaling Policy: `make update-scaling-policy`
```bash
terraform apply -auto-approve -var=scaling_policy_type=step_scaling
  └─ Switches between target_tracking and step_scaling
     (Used during Experiment 2: Autoscaling Analysis)
```

---

### Update Server Code: `make update-server`
```bash
python scripts/update_ecr_ecs.py
  1. Build Docker image from flash-sale-platform/src
  2. Tag image with ECR URL
  3. Login to ECR (aws ecr get-login-password)
  4. Push image to ECR
  5. Force ECS service deployment (--force-new-deployment)
  6. Wait for service to stabilize
```

---

### Restart ECS Service: `make restart-server`
```bash
python scripts/restart_ecs.py
  └─ Triggers ECS service restart without code changes
```

---

## Phase 2: Load Testing

### Test Entry Point
```bash
make test-thundering-herd-target
```

## Orchestration: `run_scenario.py`

**1. Load Configuration from `scenarios.yaml`**
```yaml
thundering_herd_target_tracking:
  workers: 4
  users: 500
  spawn_rate: 500
  run_time: 180
  user_class: "AggressiveBuyer"
  scaling_policy: "target_tracking"
```

**2. Configure Autoscaling Policy**
```python
_configure_scaling_policy(env, 'target_tracking')
  ├─ Run: terraform apply -var=scaling_policy_type=target_tracking
  ├─ Wait for ECS service to stabilize
  └─ 30s buffer before test starts
```

**3. Build Environment Variables**
```python
env['API_HOST'] = ALB_DNS_NAME (from .env)
env['USER_CLASS'] = 'AggressiveBuyer'
env['USERS'] = 500
env['SPAWN_RATE'] = 500
env['RUN_TIME'] = 180
env['WORKERS'] = 4
env['TEST_NAME'] = 'thundering_herd_target_tracking'
env['TEST_RESULTS_FILE_NAME'] = 'thundering_herd_target_tracking_test_results.json'
```

**4. Launch Docker Compose**
```bash
docker-compose up -d --build
  ├─ Spins up locust-master container
  └─ Spins up 4 locust-worker containers
```

**5. Wait for Test Completion**
```python
time.sleep(run_time + 10)  # 180s + 10s buffer
```

**6. Cleanup**
```bash
docker-compose down
```

---

## Locust Execution: `locustfile.py`

**User Class Selection** (based on `USER_CLASS` env var)
```python
user_class = os.getenv('USER_CLASS', 'NormalUser')
```

### AggressiveBuyer Tasks (for thundering_herd)
```
@task(30) browse_product   → GET /products/{1-3}
@task(50) update_product   → PUT /products/{1-3} (simulates inventory decrement)
@task(5)  create_product   → POST /products (simulates order creation)
@task(10) browse_catalog   → GET /products?ids=1,2,3
```

### NormalUser Tasks (for baseline)
```
@task(35) browse_product   → GET /products/{1-50}
@task(30) browse_catalog   → GET /products?ids=...
@task(20) create_product   → POST /products
@task(5)  update_product   → PUT /products/{1-50}
@task(10) health_check     → GET /health
```

### ChaosTestUser Tasks (for chaos testing)
```
@task(25) browse_product
@task(15) create_product
@task(10) update_product
@task(20) browse_catalog
@task(15) health_check
```

**Request Flow per Task**
```python
start_time = datetime.now()
→ HTTP request (with catch_response=True)
→ end_time = datetime.now()
→ Calculate response_time in ms
→ Append to test_results[] with:
    - operation name
    - user_class
    - response_time
    - success (status_code check)
    - status_code
    - timestamp
```

---

## Data Collection Lifecycle

**Test Start Event** (`@events.test_start`)
```python
metrics_collector.start_collection()
  └─ Starts CloudWatch polling in background thread
```

**During Test** (every request)
```python
test_results.append({...})  # In-memory list
```

**Test Stop Event** (`@events.test_stop`)
```python
1. metrics_collector.stop_collection()
2. Save test_results to JSON:
   → results/formatted_locust_data/{TEST_NAME}_test_results.json
   → Includes summary statistics per operation
3. time.sleep(120)  # Wait for CloudWatch to finalize
4. metrics_collector.export_all_metrics(format='json')
   → results/cloudwatch_metrics/{TEST_NAME}_cloudwatch_metrics.json
```

---

## Analysis: `analyze_baseline_thundering_herd.py`

**Example: `analyze_baseline_thundering_herd.py`**

**Data Loading**
```python
Load: baseline_test_results.json
Load: thundering_herd_target_tracking_test_results.json
```

**Statistical Processing**
```python
For each operation:
  ├─ Calculate P50 response time
  ├─ Calculate average response time
  ├─ Calculate error_rate = (failures / total) * 100
  └─ Count total requests
```

**Visualization Output**
```python
1. P50 Response Time by Operation (side-by-side bars)
2. Error Rate by Operation (side-by-side bars)
3. Response Time Over Time (line charts, 10s buckets)
   └─ Saved as: response_time_over_time.png
```

**Console Output**
```
- Test Configuration Summary Table
- User Class Behaviors
- P50 Response Time Summary (pivot table)
- Average Response Time Summary
- Error Rate Summary
- Request Count by Operation
```

---

## Phase 4: Cleanup

### Full Cleanup: `make clean`
```bash
1. docker-compose down              # Stop Locust containers
2. docker system prune -f           # Remove unused Docker resources
3. python scripts/destroy_aws.py    # Destroy all AWS infrastructure
   └─ terraform destroy -auto-approve
```

**What gets destroyed:**
- ECS Cluster & Services
- Application Load Balancer
- RDS Database
- Redis Cache
- RabbitMQ
- Security Groups
- VPC Resources
- CloudWatch Log Groups

---

## Output Structure
```
results/
├─ formatted_locust_data/
│  ├─ baseline_test_results.json
│  ├─ thundering_herd_target_tracking_test_results.json
│  └─ ...
├─ cloudwatch_metrics/
│  ├─ baseline_cloudwatch_metrics.json
│  └─ thundering_herd_target_tracking_cloudwatch_metrics.json
├─ raw_locust_data/
│  └─ {TEST_NAME}_stats.csv (Locust CSV output)
├─ locust_reports/
│  └─ {TEST_NAME}.html (Locust HTML report)
└─ charts/
   └─ response_time_over_time.png
```

---

## Key Files

| File | Purpose |
|------|---------|
| `Makefile` | Entry points for all test scenarios |
| `scenarios.yaml` | Test configurations (users, spawn_rate, user_class, scaling_policy) |
| `run_scenario.py` | Orchestrates Terraform, Docker Compose, and test execution |
| `locustfile.py` | Defines user behavior classes and request patterns |
| `docker-compose.yaml` | Configures Locust master + workers with volume mounts |
| `analyze_baseline_thundering_herd.py` | Generates comparison visualizations |
| `.env` | Stores AWS resource names (ALB_DNS_NAME, ECS_SERVICE_NAME, etc.) |

---

## Running Experiments

### Experiment 1: High-Contention Inventory
```bash
make run-exp1
  ├─ make test-baseline          # Establish baseline
  └─ make test-high-contention   # 100 users fight for 3 products
```

### Experiment 2: Autoscaling Analysis
```bash
make run-exp2
  ├─ make test-baseline          # Establish baseline
  ├─ make test-thundering-herd   # 500 user spike
  └─ make test-sustained         # Verify stability after scale-up
```

### Experiment 3: Failure Recovery
```bash
make run-exp3
  ├─ make test-baseline          # Establish baseline
  └─ make test-chaos             # Run with manual failure injection
```
**Manual Step:** Inject failures during test (stop ECS tasks, kill containers)

### Experiment 4: Fairness Evaluation
```bash
make run-exp4
  ├─ make test-baseline          # Establish baseline
  └─ make test-fairness          # 300 users, 30 items (10x demand)
```
**Note:** Run twice - with/without FIFO queue configuration

### Run All Experiments
```bash
make run-all
  ├─ make run-exp1
  ├─ make run-exp2
  ├─ make run-exp3
  └─ make run-exp4
```

---

## Phase 3: Analysis

### Analyze Results: `make analyze`
```bash
python scripts/analyze_results.py
```

**Example: `analyze_baseline_thundering_herd.py`**