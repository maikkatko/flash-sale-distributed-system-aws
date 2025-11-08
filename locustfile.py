import time
from locust import FastHttpUser, task, between, events
from scripts.metrics_collector import metrics_collector


class NormalUser(FastHttpUser):
    """For: baseline, sustained_load"""
    wait_time = between(1, 3)
    
    @task(30)
    def browse_product(self):
        self.client.get("/products/1")
    
    @task(50)
    def checkout(self):
        self.client.post("/checkout", json={
            "product_id": 1,
            "quantity": 1,
            "user_id": f"user_{id(self)}"
        })
    
    @task(20)
    def browse_catalog(self):
        self.client.get("/products?ids=1,2,3")


class AggressiveBuyer(FastHttpUser):
    """For: high_contention, thundering_herd, fairness_10x_demand"""
    wait_time = between(0.1, 0.5)
    
    @task(5)
    def browse_product(self):
        self.client.get("/products/1")
    
    @task(90)
    def checkout(self):
        self.client.post("/checkout", json={
            "product_id": 1,
            "quantity": 1,
            "user_id": f"user_{id(self)}"
        })
    
    @task(5)
    def check_inventory(self):
        self.client.get("/inventory/1")


class ChaosTestUser(FastHttpUser):
    """For: chaos_baseline - needs consistent traffic to observe failures"""
    wait_time = between(0.5, 2)
    
    @task(40)
    def browse_product(self):
        self.client.get("/products/1")
    
    @task(50)
    def checkout(self):
        self.client.post("/checkout", json={
            "product_id": 1,
            "quantity": 1,
            "user_id": f"user_{id(self)}"
        })
    
    @task(10)
    def health_check(self):
        # Helps detect when services go down
        self.client.get("/health")

@events.test_start.add_listener
def on_test_start(environment, **kwargs):
    metrics_collector.start_collection()


@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    metrics_collector.stop_collection()
    
    print("[METRICS] Waiting 2 minutes for CloudWatch to finalize data...")
    time.sleep(120)
    
    metrics_collector.export_all_metrics(output_format='json')