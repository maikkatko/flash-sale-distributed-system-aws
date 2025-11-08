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
