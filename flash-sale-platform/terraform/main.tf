# Wire together modules: network, ecr, logging, alb, ecs

module "network" {
  source         = "./modules/network"
  service_name   = var.service_name
  container_port = var.container_port
}

module "ecr" {
  source          = "./modules/ecr"
  repository_name = var.ecr_repository_name
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

# NEW: Application Load Balancer
module "alb" {
  source                 = "./modules/alb"
  service_name           = var.service_name
  vpc_id                 = module.network.vpc_id
  subnet_ids             = module.network.subnet_ids
  alb_security_group_id  = module.network.alb_security_group_id
  container_port         = var.container_port
}

# UPDATED: ECS with Auto-Scaling and ALB Integration
module "ecs" {
  source             = "./modules/ecs"
  service_name       = var.service_name
  image              = "${module.ecr.repository_url}:latest"
  container_port     = var.container_port
  subnet_ids         = module.network.subnet_ids
  security_group_ids = [module.network.security_group_id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  log_group_name     = module.logging.log_group_name
  region             = var.aws_region
  
  # NEW: ALB integration
  target_group_arn   = module.alb.target_group_arn
  
  # NEW: Auto-scaling parameters
  min_capacity       = var.min_capacity
  max_capacity       = var.max_capacity
  cpu_target_value   = var.cpu_target_value
  scale_out_cooldown = var.scale_out_cooldown
  scale_in_cooldown  = var.scale_in_cooldown
}

# Build & push the Go app image into ECR
resource "docker_image" "app" {
  name = "${module.ecr.repository_url}:latest"
  build {
    context    = "../src/products"
    platform   = "linux/amd64"
    no_cache   = true
  }
}

resource "docker_registry_image" "app" {
  name = docker_image.app.name

  depends_on = [
    module.ecr
  ]
}