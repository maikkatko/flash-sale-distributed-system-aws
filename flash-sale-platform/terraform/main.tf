# Wire together modules: network, ecr, logging, alb, ecs

module "network" {
  source         = "./modules/network"
  service_name   = var.service_name
}

module "ecr_products" {
  source          = "./modules/ecr"
  repository_name = "${var.service_name}-products"
}

module "ecr_orders" {
  source          = "./modules/ecr"
  repository_name = "${var.service_name}-orders"
}

module "logging" {
  source            = "./modules/logging"
  service_name      = var.service_name
  retention_in_days = var.log_retention_days
}

# Reuse existing IAM role for ECS tasks
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

# --- Security Groups ---

resource "aws_security_group" "alb" {
  name        = "${var.service_name}-alb-sg"
  description = "Allow HTTP traffic to ALB"
  vpc_id      = module.network.vpc_id

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "ecs_tasks" {
  name        = "${var.service_name}-ecs-tasks-sg"
  description = "Allow traffic to the ECS tasks"
  vpc_id      = module.network.vpc_id

  # Allow inbound traffic from the ALB
  ingress {
    from_port       = var.container_port_products
    to_port         = var.container_port_products
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  # Allow inbound traffic from the ALB for orders service
  ingress {
    from_port       = var.container_port_orders
    to_port         = var.container_port_orders
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  # Allow all outbound traffic for pulling images and other needs
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# --- SQS ---
module "sqs" {
  source      = "./modules/sqs"
  name_prefix = "${var.service_name}-orders"
  tags = {
    Project = var.service_name
  }
}

module "alb" {
  source                 = "./modules/alb"
  service_name           = var.service_name
  vpc_id                 = module.network.vpc_id
  subnet_ids             = module.network.public_subnet_ids
  alb_security_group_id  = aws_security_group.alb.id
  # Target group for products service
  products_container_port = var.container_port_products
  # Target group for orders service
  orders_container_port = var.container_port_orders
  # Health check path for orders service
  health_check_path = "/health"
}

# --- Database ---

module "rds" {
  source                = "./modules/rds"
  service_name          = var.service_name
  vpc_id                = module.network.vpc_id
  private_subnet_ids    = module.network.private_subnet_ids
  ecs_security_group_id = aws_security_group.ecs_tasks.id
  db_name               = var.db_name
  db_username           = var.db_username
  db_password           = var.db_password
}

# UPDATED: ECS with Auto-Scaling and ALB Integration
module "ecs_products" {
  source             = "./modules/ecs"
  service_name       = var.service_name
  image              = docker_registry_image.products_app_registry.name
  container_port     = var.container_port_products
  subnet_ids         = module.network.private_subnet_ids
  security_group_ids = [aws_security_group.ecs_tasks.id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  log_group_name     = module.logging.log_group_name
  region             = var.aws_region
  alb_arn_suffix     = module.alb.alb_arn_suffix
  target_group_arn   = module.alb.products_target_group_arn

  # Scaling config
  min_capacity        = var.min_capacity
  max_capacity        = var.max_capacity
  scaling_policy_type = var.scaling_policy_type
  step_scaling_config = var.step_scaling_config
  cpu_target_value    = var.cpu_target_value
  scale_out_cooldown  = var.scale_out_cooldown
  scale_in_cooldown   = var.scale_in_cooldown

  # Pass database connection details as environment variables
  environment_variables = {
    DB_HOST     = module.rds.db_endpoint
    DB_NAME     = module.rds.db_name
    DB_USER     = var.db_username
    DB_PASSWORD = var.db_password
    PORT        = var.container_port_products
  }
}

module "ecs_orders" {
  source             = "./modules/ecs"
  service_name       = "${var.service_name}-orders"
  image              = docker_registry_image.orders_app_registry.name
  container_port     = var.container_port_orders
  subnet_ids         = module.network.private_subnet_ids
  security_group_ids = [aws_security_group.ecs_tasks.id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  log_group_name     = module.logging.log_group_name
  region             = var.aws_region
  alb_arn_suffix     = module.alb.alb_arn_suffix
  target_group_arn   = module.alb.orders_target_group_arn

  # Scaling config
  min_capacity        = var.min_capacity
  max_capacity        = var.max_capacity
  scaling_policy_type = var.scaling_policy_type
  step_scaling_config = var.step_scaling_config
  cpu_target_value    = var.cpu_target_value
  scale_out_cooldown  = var.scale_out_cooldown
  scale_in_cooldown   = var.scale_in_cooldown

  # Pass database and SQS connection details as environment variables
  environment_variables = {
    DB_HOST       = module.rds.db_endpoint
    DB_NAME       = module.rds.db_name
    DB_USER       = var.db_username
    DB_PASSWORD   = var.db_password
    PORT          = var.container_port_orders
    SQS_QUEUE_URL = module.sqs.order_queue_url
    AWS_REGION    = var.aws_region
    REDIS_ADDR    = "dummy:6379" # Placeholder, will be replaced by ElastiCache module
  }
  depends_on = [module.sqs]
}

# Build & push the Go app image into ECR
resource "docker_image" "products_app" {
  name  = "${module.ecr_products.repository_url}:latest"
  build {
    context    = "../src/products"
    platform   = "linux/amd64"
  }
  depends_on = [module.ecr_products]
}

resource "docker_registry_image" "products_app_registry" {
  name = docker_image.products_app.name
}

resource "docker_image" "orders_app" {
  name  = "${module.ecr_orders.repository_url}:latest"
  build {
    context  = "../src/orders"
    platform = "linux/amd64"
  }
  depends_on = [module.ecr_orders]
}

resource "docker_registry_image" "orders_app_registry" {
  name = docker_image.orders_app.name
}