variable "service_name" {
  description = "Name of the ECS service"
  type        = string
}

variable "image" {
  description = "Docker image to run"
  type        = string
}

variable "container_port" {
  description = "Port the container listens on"
  type        = number
}

variable "subnet_ids" {
  description = "Subnet IDs for ECS tasks"
  type        = list(string)
}

variable "security_group_ids" {
  description = "Security group IDs for ECS tasks"
  type        = list(string)
}

variable "execution_role_arn" {
  description = "ECS task execution role ARN"
  type        = string
}

variable "task_role_arn" {
  description = "ECS task role ARN"
  type        = string
}

variable "log_group_name" {
  description = "CloudWatch log group name"
  type        = string
}

variable "region" {
  description = "AWS region"
  type        = string
}

# NEW: ALB Integration
variable "target_group_arn" {
  description = "ALB target group ARN"
  type        = string
}

# NEW: Auto-scaling parameters
variable "min_capacity" {
  description = "Minimum number of tasks"
  type        = number
  default     = 2
}

variable "max_capacity" {
  description = "Maximum number of tasks"
  type        = number
  default     = 4
}

variable "cpu_target_value" {
  description = "Target CPU utilization percentage for auto-scaling"
  type        = number
  default     = 70
}

variable "scale_out_cooldown" {
  description = "Cooldown period (seconds) after scale-out"
  type        = number
  default     = 300
}

variable "scale_in_cooldown" {
  description = "Cooldown period (seconds) after scale-in"
  type        = number
  default     = 300
}