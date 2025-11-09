import os
import json
import random
import time
from datetime import datetime
from locust import HttpUser, task, between, events
from scripts.metrics_collector import metrics_collector

user_class = os.getenv('USER_CLASS', 'NormalUser')
 
test_results = []


class NormalUser(HttpUser):
    """For: baseline, sustained_load"""
    weight = 1 if user_class == 'NormalUser' else 0
    host = os.getenv('API_HOST')
    wait_time = between(1, 3)
    
    @task(40)
    def browse_product(self):
        start_time = datetime.now()
        
        product_id = random.randint(1, 5)
        
        with self.client.get(
            f"/products/{product_id}",
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
            
            test_results.append({
                "operation": "browse_product",
                "user_class": "NormalUser",
                "response_time": round(response_time, 2),
                "success": response.status_code == 200,
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })
    
    @task(30)
    def browse_catalog(self):
        start_time = datetime.now()
        
        with self.client.get(
            "/products?ids=1,2,3",
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
            
            test_results.append({
                "operation": "browse_catalog",
                "user_class": "NormalUser",
                "response_time": round(response_time, 2),
                "success": response.status_code == 200,
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })
    
    @task(20)
    def create_product(self):
        """Simulates checkout write load"""
        start_time = datetime.now()
        
        with self.client.post(
            "/products",
            json={
                "name": f"Order-{random.randint(10000, 99999)}",
                "description": "Simulated order write",
                "price": round(random.uniform(10, 500), 2),
                "stock": random.randint(1, 10)
            },
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
            
            test_results.append({
                "operation": "create_product",
                "user_class": "NormalUser",
                "response_time": round(response_time, 2),
                "success": response.status_code == 201,
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })
    
    @task(10)
    def health_check(self):
        start_time = datetime.now()
        
        with self.client.get(
            "/health",
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
            
            test_results.append({
                "operation": "health_check",
                "user_class": "NormalUser",
                "response_time": round(response_time, 2),
                "success": response.status_code == 200,
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })


class AggressiveBuyer(HttpUser):
    """For: high_contention, thundering_herd, fairness_10x_demand"""
    weight = 1 if user_class == 'AggressiveBuyer' else 0
    host = os.getenv('API_HOST')
    wait_time = between(0.1, 0.5)
    
    @task(30)
    def browse_product(self):
        start_time = datetime.now()
        
        # High contention on flash sale items (1-3)
        product_id = random.randint(1, 3)
        
        with self.client.get(
            f"/products/{product_id}",
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
            
            test_results.append({
                "operation": "browse_product",
                "user_class": "AggressiveBuyer",
                "response_time": round(response_time, 2),
                "success": response.status_code == 200,
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })
    
    @task(50)
    def update_product(self):
        """Simulates inventory decrement under high contention"""
        start_time = datetime.now()
        
        # All buyers fight over same 3 products
        product_id = random.randint(1, 3)
        
        with self.client.put(
            f"/products/{product_id}",
            json={
                "name": f"Flash Sale Item {product_id}",
                "description": "High demand item",
                "price": 499.99,
                "stock": random.randint(0, 5)  # Simulate inventory changes
            },
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
            
            test_results.append({
                "operation": "update_product",
                "user_class": "AggressiveBuyer",
                "response_time": round(response_time, 2),
                "success": response.status_code == 204,
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })
    
    @task(15)
    def create_product(self):
        """Simulates checkout write load"""
        start_time = datetime.now()
        
        with self.client.post(
            "/products",
            json={
                "name": f"Order-{random.randint(10000, 99999)}",
                "description": "Aggressive purchase",
                "price": round(random.uniform(100, 1000), 2),
                "stock": 1
            },
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
            
            test_results.append({
                "operation": "create_product",
                "user_class": "AggressiveBuyer",
                "response_time": round(response_time, 2),
                "success": response.status_code == 201,
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })
    
    @task(5)
    def browse_catalog(self):
        start_time = datetime.now()
        
        with self.client.get(
            "/products?ids=1,2,3",
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
            
            test_results.append({
                "operation": "browse_catalog",
                "user_class": "AggressiveBuyer",
                "response_time": round(response_time, 2),
                "success": response.status_code == 200,
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })


class ChaosTestUser(HttpUser):
    """For: chaos_baseline - needs consistent traffic to observe failures"""
    weight = 1 if user_class == 'ChaosTestUser' else 0
    host = os.getenv('API_HOST')
    wait_time = between(0.5, 2)
    
    @task(40)
    def browse_product(self):
        start_time = datetime.now()
        
        product_id = random.randint(1, 5)
        
        with self.client.get(
            f"/products/{product_id}",
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
            
            test_results.append({
                "operation": "browse_product",
                "user_class": "ChaosTestUser",
                "response_time": round(response_time, 2),
                "success": response.status_code == 200,
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })
    
    @task(25)
    def update_product(self):
        """Simulates writes during chaos"""
        start_time = datetime.now()
        
        product_id = random.randint(1, 5)
        
        with self.client.put(
            f"/products/{product_id}",
            json={
                "name": f"Product {product_id}",
                "description": "Chaos test",
                "price": 99.99,
                "stock": random.randint(5, 50)
            },
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
            
            test_results.append({
                "operation": "update_product",
                "user_class": "ChaosTestUser",
                "response_time": round(response_time, 2),
                "success": response.status_code == 204,
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })
    
    @task(20)
    def browse_catalog(self):
        start_time = datetime.now()
        
        with self.client.get(
            "/products?ids=1,2,3,4,5",
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
            
            test_results.append({
                "operation": "browse_catalog",
                "user_class": "ChaosTestUser",
                "response_time": round(response_time, 2),
                "success": response.status_code == 200,
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })
    
    @task(15)
    def health_check(self):
        start_time = datetime.now()
        
        with self.client.get(
            "/health",
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
            
            test_results.append({
                "operation": "health_check",
                "user_class": "ChaosTestUser",
                "response_time": round(response_time, 2),
                "success": response.status_code == 200,
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })


@events.test_start.add_listener
def on_test_start(environment, **kwargs):
    """Start CloudWatch metrics collection when test begins"""
    print(f"\n{'='*60}")
    print("Starting CloudWatch metrics collection...")
    print(f"{'='*60}\n")
    metrics_collector.start_collection()


@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    """Save results and export CloudWatch metrics when test completes"""
    
    # Stop metrics collection first
    print(f"\n{'='*60}")
    print("Stopping CloudWatch metrics collection...")
    print(f"{'='*60}\n")
    metrics_collector.stop_collection()
    
    # Save locust test results
    print(f"\n{'='*60}")
    print("Saving test results...")
    print(f"Total operations: {len(test_results)}")
    print(f"{'='*60}\n")

    if test_results:
        test_name = os.getenv('TEST_NAME', 'test')
        filename = os.getenv('TEST_RESULTS_FILE_NAME', f'{test_name}_results.json')
        filepath = os.path.join("/mnt/locust/results/formatted_locust_data", filename)

        os.makedirs(os.path.dirname(filepath), exist_ok=True)

        operations = {}
        for result in test_results:
            op = result['operation']
            if op not in operations:
                operations[op] = {
                    'count': 0,
                    'success': 0,
                    'failed': 0,
                    'total_response_time': 0
                }
            operations[op]['count'] += 1
            if result['success']:
                operations[op]['success'] += 1
            else:
                operations[op]['failed'] += 1
            operations[op]['total_response_time'] += result['response_time']

        summary = {
            'test_name': test_name,
            'total_requests': len(test_results),
            'operations': {
                op: {
                    'count': stats['count'],
                    'success_rate': round(stats['success'] / stats['count'] * 100, 2),
                    'avg_response_time': round(stats['total_response_time'] / stats['count'], 2)
                }
                for op, stats in operations.items()
            }
        }

        output = {
            'summary': summary,
            'results': test_results
        }

        with open(filepath, 'w', encoding='utf-8') as f:
            json.dump(output, f, indent=2)

        print(f"Successfully saved results to {filepath}")
    else:
        print("WARNING: test_results is empty!")
    
    # Wait for CloudWatch to finalize data
    print(f"\n{'='*60}")
    print("[METRICS] Waiting 2 minutes for CloudWatch to finalize data...")
    print(f"{'='*60}\n")
    time.sleep(120)
    
    # Export CloudWatch metrics
    print(f"\n{'='*60}")
    print("Exporting CloudWatch metrics...")
    print(f"{'='*60}\n")
    metrics_collector.export_all_metrics(output_format='json')
    
    print(f"\n{'='*60}")
    print("Test complete! All data saved.")
    print(f"{'='*60}\n")