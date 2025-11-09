output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = aws_lb.this.dns_name
}

output "alb_arn" {
  description = "ARN of the Application Load Balancer"
  value       = aws_lb.this.arn
}

output "target_group_arn" {
  description = "ARN of the target group"
  value       = aws_lb_target_group.this.arn
}

output "alb_zone_id" {
  description = "Zone ID of the ALB"
  value       = aws_lb.this.zone_id
}

output "alb_arn_suffix" {
  description = "ALB ARN suffix for metrics"
  value       = split("loadbalancer/", aws_lb.this.arn)[1]
}