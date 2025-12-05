output "primary_endpoint_address" {
  description = "The connection endpoint for the primary Redis node. Includes port."
  value       = aws_elasticache_replication_group.main.primary_endpoint_address
}

output "security_group_id" {
  description = "The ID of the security group for the Redis cluster."
  value       = aws_security_group.redis.id
}
