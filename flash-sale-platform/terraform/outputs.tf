# /terraform/outputs.tf

output "alb_dns_name" {
  description = "The DNS name of the Application Load Balancer."
  value       = module.alb.alb_dns_name
}

output "alb_url" {
  description = "Full URL to access the service"
  value       = "http://${module.alb.alb_dns_name}"
}