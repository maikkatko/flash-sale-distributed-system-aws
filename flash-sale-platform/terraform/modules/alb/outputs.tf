output "alb_dns_name" {
  description = "The DNS name of the Application Load Balancer."
  value       = aws_lb.this.dns_name
}

output "products_target_group_arn" {
  description = "ARN of the target group for the products service"
  value       = aws_lb_target_group.products.arn
}

output "orders_target_group_arn" {
  description = "ARN of the target group for the orders service"
  value       = aws_lb_target_group.orders.arn
}

output "alb_arn_suffix" {
  description = "ARN suffix of the ALB"
  value       = aws_lb.this.arn_suffix
}