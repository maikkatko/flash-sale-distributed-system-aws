variable "service_name" {
  description = "The name of the service."
  type        = string
}

variable "image" {
  description = "The Docker image to use for the task."
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

variable "target_group_arn" {
  description = "ALB target group ARN"
  type        = string
}

variable "min_capacity" {
  description = "Minimum number of tasks for auto-scaling."
  type        = number
}

variable "max_capacity" {
  description = "Maximum number of tasks for auto-scaling."
  type        = number
}

variable "cpu_target_value" {
  description = "Target CPU utilization for scaling."
  type        = number
}

variable "scale_in_cooldown" { type = number }
variable "scale_out_cooldown" { type = number }

variable "environment_variables" {
  description = "A map of environment variables to pass to the container."
  type        = map(string)
  default     = {}
}