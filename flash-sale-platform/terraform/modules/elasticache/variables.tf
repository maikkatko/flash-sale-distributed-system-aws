variable "service_name" {
  description = "The name of the service, used for resource naming."
  type        = string
}

variable "vpc_id" {
  description = "The ID of the VPC where the ElastiCache cluster will be deployed."
  type        = string
}

variable "private_subnet_ids" {
  description = "A list of private subnet IDs for the ElastiCache subnet group."
  type        = list(string)
}

variable "ecs_security_group_id" {
  description = "The security group ID of the ECS tasks that need to access Redis."
  type        = string
}

variable "instance_type" {
  description = "The instance type for the Redis nodes."
  type        = string
  default     = "cache.t3.micro"
}

variable "tags" {
  description = "A map of tags to assign to the resources."
  type        = map(string)
  default     = {}
}
