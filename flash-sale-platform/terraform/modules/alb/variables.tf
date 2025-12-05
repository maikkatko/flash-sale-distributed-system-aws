variable "service_name" {
  description = "Name of the service"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID where ALB will be created"
  type        = string
}

variable "subnet_ids" {
  description = "List of subnet IDs for ALB"
  type        = list(string)
}

variable "alb_security_group_id" {
  description = "Security group ID for ALB"
  type        = string
}

variable "products_container_port" {
  description = "Port the products container listens on"
  type        = number
}

variable "orders_container_port" {
  description = "Port the orders container listens on"
  type        = number
}

variable "health_check_path" {
  description = "Health check path for the target groups"
  type        = string
  default     = "/"
}