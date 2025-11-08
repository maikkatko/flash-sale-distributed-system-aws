import os
import json
import random
import time
from datetime import datetime
from locust import HttpUser, task, between, events
from scripts.metrics_collector import metrics_collector
 
test_results = []


class NormalUser(HttpUser):
    """For: baseline, sustained_load"""
    host = os.getenv('API_HOST')
    wait_time = between(1, 3)
    
    @task(30)
    def browse_product(self):
        start_time = datetime.now()
        
        with self.client.get(
            "/products/1",
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
    
    @task(50)
    def checkout(self):
        start_time = datetime.now()
        
        with self.client.post(
            "/checkout",
            json={
                "product_id": 1,
                "quantity": 1,
                "user_id": f"user_{id(self)}"
            },
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
            
            test_results.append({
                "operation": "checkout",
                "user_class": "NormalUser",
                "response_time": round(response_time, 2),
                "success": response.status_code in [200, 201],
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })
    
    @task(20)
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


class AggressiveBuyer(HttpUser):
    """For: high_contention, thundering_herd, fairness_10x_demand"""
    host = os.getenv('API_HOST')
    wait_time = between(0.1, 0.5)
    
    @task(5)
    def browse_product(self):
        start_time = datetime.now()
        
        with self.client.get(
            "/products/1",
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
    
    @task(90)
    def checkout(self):
        start_time = datetime.now()
        
        with self.client.post(
            "/checkout",
            json={
                "product_id": 1,
                "quantity": 1,
                "user_id": f"user_{id(self)}_{random.randint(1, 10000)}"
            },
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
            
            test_results.append({
                "operation": "checkout",
                "user_class": "AggressiveBuyer",
                "response_time": round(response_time, 2),
                "success": response.status_code in [200, 201],
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })
    
    @task(5)
    def check_inventory(self):
        start_time = datetime.now()
        
        with self.client.get(
            "/inventory/1",
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
            
            test_results.append({
                "operation": "check_inventory",
                "user_class": "AggressiveBuyer",
                "response_time": round(response_time, 2),
                "success": response.status_code == 200,
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })


class ChaosTestUser(HttpUser):
    """For: chaos_baseline - needs consistent traffic to observe failures"""
    host = os.getenv('API_HOST')
    wait_time = between(0.5, 2)
    
    @task(40)
    def browse_product(self):
        start_time = datetime.now()
        
        with self.client.get(
            "/products/1",
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
    
    @task(50)
    def checkout(self):
        start_time = datetime.now()
        
        with self.client.post(
            "/checkout",
            json={
                "product_id": 1,
                "quantity": 1,
                "user_id": f"user_{id(self)}"
            },
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
            
            test_results.append({
                "operation": "checkout",
                "user_class": "ChaosTestUser",
                "response_time": round(response_time, 2),
                "success": response.status_code in [200, 201],
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
        filepath = os.path.join("/mnt/locust/results/data", filename)

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