output "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  value       = var.service_name
}

output "ecs_service_name" {
  description = "Name of the ECS service"
  value       = var.service_name
}

# NEW: ALB DNS name - This is your new endpoint!
output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer (use this for testing)"
  value       = module.alb.alb_dns_name
}

output "alb_url" {
  description = "Full URL to access the service"
  value       = "http://${module.alb.alb_dns_name}"
}

output "target_group_arn" {
  description = "Target group ARN"
  value       = module.alb.target_group_arn
}