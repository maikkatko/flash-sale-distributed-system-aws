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

variable "container_port_orders" {
  description = "The port the orders container will listen on."
  type        = number
  default     = 8082
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
  default     = "flashsale"  # Add this
}

variable "db_username" {
  description = "The master username for the database."
  type        = string
  sensitive   = true
  default     = "admin"  # Add this
}

variable "db_password" {
  description = "The master password for the database."
  type        = string
  sensitive   = true
  default     = "SecurePassword123!"  # Add this
}

variable "scaling_policy_type" {
  description = "Scaling policy: target_tracking or step_scaling"
  type        = string
  default     = "target_tracking"
  
  validation {
    condition     = contains(["target_tracking", "step_scaling"], var.scaling_policy_type)
    error_message = "Must be target_tracking or step_scaling"
  }
}

variable "step_scaling_config" {
  description = "Step scaling configuration"
  type = object({
    metric_aggregation_type = string
    adjustment_type         = string
    cooldown                = number
    steps = list(object({
      scaling_adjustment          = number
      metric_interval_lower_bound = number
      metric_interval_upper_bound = number
    }))
  })
  default = {
    metric_aggregation_type = "Average"
    adjustment_type         = "ChangeInCapacity"
    cooldown                = 60
    steps = [
      {
        scaling_adjustment          = 2
        metric_interval_lower_bound = 0
        metric_interval_upper_bound = 500
      },
      {
        scaling_adjustment          = 5
        metric_interval_lower_bound = 500
        metric_interval_upper_bound = 1000
      },
      {
        scaling_adjustment          = 10
        metric_interval_lower_bound = 1000
        metric_interval_upper_bound = null
      }
    ]
  }
}