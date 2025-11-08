import os
import json
import random
from datetime import datetime
from locust import FastHttpUser, task, between, events

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


@events.test_stop.add_listener
def on_test_stopping(environment, **kwargs):
    """Save results when test completes - runs on workers"""
    print(f"\n{'='*60}")
    print("IN ON_TEST_STOPPING")
    print(f"test_results length: {len(test_results)}")
    print(f"Creates: {CREATE_COUNT}, Adds: {ADD_COUNT}, Gets: {GET_COUNT}")
    print(f"{'='*60}\n")

    if not test_results:
        print("WARNING: test_results is empty!")
        return

    filename = os.getenv('TEST_RESULTS_FILE_NAME', 'mysql_test_results.json')
    filepath = os.path.join("/mnt/locust/results/db_data", filename)

    print(f"About to write {len(test_results)} results to {filepath}")

    with open(filepath, 'w', encoding='utf-8') as f:
        json.dump(test_results, f, indent=2)

    print(f"Successfully saved to {filepath}")
