# Region to deploy into
variable "aws_region" {
  type    = string
  default = "us-east-1"
}

# ECR & ECS settings
variable "ecr_repository_name" {
  type    = string
  default = "flash-sale"
}

variable "service_name" {
  type    = string
  default = "flash-sale"
}

variable "container_port_products" {
  type    = number
  default = 8081
}

variable "container_port" {
  description = "The port the container listens on. Used by network and ALB modules."
  type        = number
  default     = 8080
}

# How long to keep logs
variable "log_retention_days" {
  type    = number
  default = 7
}

# NEW: Auto-scaling configuration
variable "min_capacity" {
  description = "Minimum number of ECS tasks"
  type        = number
  default     = 1
}

variable "max_capacity" {
  description = "Maximum number of ECS tasks"
  type        = number
  default     = 15
}

variable "cpu_target_value" {
  description = "Target CPU utilization percentage for auto-scaling"
  type        = number
  default     = 60
}

variable "scale_out_cooldown" {
  description = "Cooldown period (seconds) after scaling out"
  type        = number
  default     = 60
}

variable "scale_in_cooldown" {
  description = "Cooldown period (seconds) after scaling in"
  type        = number
  default     = 300
}

# Database variables
variable "db_name" {
  description = "The name of the database to create."
  type        = string
}

variable "db_username" {
  description = "The master username for the database."
  type        = string
  sensitive   = true
}

variable "db_password" {
  description = "The master password for the database."
  type        = string
  sensitive   = true
}