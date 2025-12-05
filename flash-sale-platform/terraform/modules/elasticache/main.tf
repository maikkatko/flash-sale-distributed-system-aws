# Security group to control access to the Redis cluster
resource "aws_security_group" "redis" {
  name        = "${var.service_name}-redis-sg"
  description = "Allow traffic to the ElastiCache Redis cluster"
  vpc_id      = var.vpc_id

  # Allow ingress from the ECS tasks on the Redis port
  ingress {
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [var.ecs_security_group_id]
  }

  # Allow all outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = var.tags
}

# ElastiCache requires a subnet group to know which subnets to deploy into
resource "aws_elasticache_subnet_group" "main" {
  name       = "${var.service_name}-redis-subnet-group"
  subnet_ids = var.private_subnet_ids
  tags       = var.tags
}

# The Redis replication group (cluster)
resource "aws_elasticache_replication_group" "main" {
  replication_group_id          = "${var.service_name}-redis"
  description                   = "Redis cluster for the flash sale platform"
  node_type                     = var.instance_type
  port                          = 6379
  automatic_failover_enabled    = false # No failover for a single node cluster
  multi_az_enabled              = false # Not needed for a single node
  num_cache_clusters         = 1
  subnet_group_name             = aws_elasticache_subnet_group.main.name
  security_group_ids            = [aws_security_group.redis.id]
  
  # Enable encryption in transit for better security
  transit_encryption_enabled = false
  apply_immediately          = true

  tags = var.tags

  # Ensure the subnet group is created before the cluster
  depends_on = [aws_elasticache_subnet_group.main]
}
