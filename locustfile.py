import os
import random
import time
from datetime import datetime
from locust import HttpUser, task, between, events
from scripts.metrics_collector import metrics_collector

user_class = os.getenv('USER_CLASS', 'NormalUser')
 
class NormalUser(HttpUser):
    """For: baseline, sustained_load"""
    weight = 1 if user_class == 'NormalUser' else 0
    host = os.getenv('API_HOST')
    wait_time = between(1, 3)
    
    @task(35)
    def browse_product(self):
        start_time = datetime.now()
        
        product_id = random.randint(1, 5)
        
        with self.client.get(
            f"/products/{product_id}",
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000

            if response.status_code >= 400:
                print(f"GET failed: {response.status_code} - {response.text}")
                response.failure(f"Got {response.status_code}")

    @task(5)
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

            if response.status_code >= 400:
                print(f"PUT failed: {response.status_code} - {response.text}")
                response.failure(f"Got {response.status_code}")
    
    @task(30)
    def browse_catalog(self):
        start_time = datetime.now()
        
        with self.client.get(
            "/products?ids=1,2,3",
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000

            if response.status_code >= 400:
                print(f"GET failed: {response.status_code} - {response.text}")
                response.failure(f"Got {response.status_code}")
    
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

            if response.status_code >= 400:
                print(f"POST failed: {response.status_code} - {response.text}")
                response.failure(f"Got {response.status_code}")
    
    @task(10)
    def health_check(self):
        start_time = datetime.now()
        
        with self.client.get(
            "/health",
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
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

            if response.status_code >= 400:
                print(f"GET failed: {response.status_code} - {response.text}")
                response.failure(f"Got {response.status_code}")
    
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

            if response.status_code >= 400:
                print(f"PUT failed: {response.status_code} - {response.text}")
                response.failure(f"Got {response.status_code}")
    
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

            if response.status_code >= 400:
                print(f"POST failed: {response.status_code} - {response.text}")
                response.failure(f"Got {response.status_code}")
            
    @task(5)
    def browse_catalog(self):
        start_time = datetime.now()
        
        with self.client.get(
            "/products?ids=1,2,3",
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000

            if response.status_code >= 400:
                print(f"GET failed: {response.status_code} - {response.text}")
                response.failure(f"Got {response.status_code}")
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

            if response.status_code >= 400:
                print(f"GET failed: {response.status_code} - {response.text}")
                response.failure(f"Got {response.status_code}")
    
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

            if response.status_code >= 400:
                print(f"PUT failed: {response.status_code} - {response.text}")
                response.failure(f"Got {response.status_code}")
    
    @task(20)
    def browse_catalog(self):
        start_time = datetime.now()
        
        with self.client.get(
            "/products?ids=1,2,3,4,5",
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000

            if response.status_code >= 400:
                print(f"GET failed: {response.status_code} - {response.text}")
                response.failure(f"Got {response.status_code}")
    
    @task(15)
    def health_check(self):
        start_time = datetime.now()
        
        with self.client.get(
            "/health",
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000

@events.test_start.add_listener
def on_test_start(environment, **kwargs):
    global metrics_collector
    from scripts.metrics_collector import METRICS_CONFIG, AWSMetricsCollector
    metrics_collector = AWSMetricsCollector(METRICS_CONFIG)
    metrics_collector.start_collection()

@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    """Save results and export CloudWatch metrics when test completes"""
    # Stop metrics collection
    print(f"\n{'='*60}")
    print("Stopping CloudWatch metrics collection...")
    print(f"{'='*60}\n")
    metrics_collector.stop_collection()
    
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