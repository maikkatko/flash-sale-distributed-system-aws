# scripts/seed_products.py
import os
import requests
from dotenv import load_dotenv

load_dotenv()

API_HOST = os.getenv('ALB_DNS_NAME')

products = [
    {"name": "Flash Sale Item 1", "description": "Limited stock PS5", "price": 499.99, "stock": 10},
    {"name": "Flash Sale Item 2", "description": "Concert Tickets", "price": 150.00, "stock": 50},
    {"name": "Flash Sale Item 3", "description": "Limited Sneakers", "price": 200.00, "stock": 25},
    {"name": "Regular Item 4", "description": "Normal product", "price": 50.00, "stock": 1000},
    {"name": "Regular Item 5", "description": "Normal product", "price": 75.00, "stock": 500},
]

for product in products:
    response = requests.post(f"{API_HOST}/products", json=product)
    if response.status_code == 201:
        print(f"✓ Created: {product['name']}")
    else:
        print(f"✗ Failed: {product['name']} - {response.status_code}")