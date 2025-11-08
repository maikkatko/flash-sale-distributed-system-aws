import os
import json
import random
from datetime import datetime
import time
from locust import FastHttpUser, task, between, events
from scripts.metrics_collector import metrics_collector

created_carts = []
unused_carts = []
test_results = []
CREATE_COUNT = 0
ADD_COUNT = 0
GET_COUNT = 0
MAX_OPERATIONS = 50

class ShoppingCartUser(FastHttpUser):
    host = os.getenv('API_HOST')
    wait_time = between(2, 2)

    @task
    def run_test_sequence(self):
        global CREATE_COUNT, ADD_COUNT, GET_COUNT

        # Check if all operations complete
        if CREATE_COUNT >= MAX_OPERATIONS and ADD_COUNT >= MAX_OPERATIONS and GET_COUNT >= MAX_OPERATIONS:
            self.environment.runner.quit()
            return

        # Do creates first
        if CREATE_COUNT < MAX_OPERATIONS:
            self._create_cart()
        # Then add items
        elif ADD_COUNT < MAX_OPERATIONS:
            self._add_items()
        # Finally get carts
        elif GET_COUNT < MAX_OPERATIONS:
            self._get_cart()

    def _create_cart(self):
        global CREATE_COUNT

        start_time = datetime.now()
        customer_id = random.randint(1, 10000)

        print(f"Creating cart for customer {customer_id}...")

        with self.client.post(
            "/shopping-carts",
            json={"customer_id": customer_id},
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000
            success = response.status_code == 201

            if success and response.json():
                cart_id = response.json().get("shopping_cart_id")
                created_carts.append(cart_id)
                unused_carts.append(cart_id)

            test_results.append({
                "operation": "create_cart",
                "response_time": round(response_time, 2),
                "success": success,
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })

            CREATE_COUNT += 1

    def _add_items(self):
        global ADD_COUNT

        if not unused_carts:
            print("No unused carts available to add items to.")
            return

        start_time = datetime.now()
        cart_id = unused_carts.pop(0)

        with self.client.post(
            f"/shopping-carts/{cart_id}/items",
            json={
                "product_id": random.randint(1, 1000),
                "quantity": random.randint(1, 5)
            },
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000

            print(f"add_items response status: {response.status_code}")

            test_results.append({
                "operation": "add_items",
                "response_time": round(response_time, 2),
                "success": response.status_code == 204,
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })

            ADD_COUNT += 1

    def _get_cart(self):
        global GET_COUNT

        if not created_carts:
            return

        start_time = datetime.now()
        cart_id = random.choice(created_carts)

        with self.client.get(
            f"/shopping-carts/{cart_id}",
            catch_response=True
        ) as response:
            end_time = datetime.now()
            response_time = (end_time - start_time).total_seconds() * 1000

            print(f"get_cart: {response.json()}")

            test_results.append({
                "operation": "get_cart",
                "response_time": round(response_time, 2),
                "success": response.status_code == 200,
                "status_code": response.status_code,
                "timestamp": start_time.isoformat() + "Z"
            })

            GET_COUNT += 1


@events.test_start.add_listener
def on_test_start(environment, **kwargs):
    metrics_collector.start_collection()


@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    metrics_collector.stop_collection()
    
    print("[METRICS] Waiting 2 minutes for CloudWatch to finalize data...")
    time.sleep(120)
    
    metrics_collector.export_all_metrics(output_format='json')