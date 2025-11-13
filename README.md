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
