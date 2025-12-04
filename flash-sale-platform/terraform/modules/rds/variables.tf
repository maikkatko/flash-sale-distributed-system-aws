# /terraform/modules/rds/variables.tf

variable "service_name" {
  description = "The name of the service, used for naming resources."
  type        = string
}

variable "vpc_id" {
  description = "The ID of the VPC where the RDS instance will be deployed."
  type        = string
}

variable "private_subnet_ids" {
  description = "A list of private subnet IDs for the RDS instance."
  type        = list(string)
}

variable "ecs_security_group_id" {
  description = "The security group ID of the ECS tasks that need to connect to the DB."
  type        = string
}

variable "db_name" {
  description = "The name of the database to create."
  type        = string
  default = "flashsale"
}

variable "db_username" {
  description = "The master username for the database."
  type        = string
  sensitive   = true
  default = "admin"
}

variable "db_password" {
  description = "The master password for the database."
  type        = string
  sensitive   = true
  default = "SecurePassword123!"
}

