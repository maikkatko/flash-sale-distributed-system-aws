# /terraform/modules/rds/main.tf

# Security group for the RDS instance
resource "aws_security_group" "rds" {
  name        = "${var.service_name}-rds-sg"
  description = "Allow traffic to the RDS instance"
  vpc_id      = var.vpc_id

  # Egress is open by default, which is fine.
  tags = {
    Name = "${var.service_name}-rds-sg"
  }
}

# Security group rule to allow inbound traffic from ECS tasks to RDS
resource "aws_security_group_rule" "rds_ingress" {
  type                     = "ingress"
  from_port                = 3306
  to_port                  = 3306
  protocol                 = "tcp"
  source_security_group_id = var.ecs_security_group_id
  security_group_id        = aws_security_group.rds.id
  description              = "Allow MySQL traffic from ECS tasks"
}

# Subnet group for RDS, placing it in private subnets
resource "aws_db_subnet_group" "rds" {
  name       = "${var.service_name}-rds-subnet-group"
  subnet_ids = var.private_subnet_ids

  tags = {
    Name = "${var.service_name}-rds-subnet-group"
  }
}

# The RDS MySQL instance
resource "aws_db_instance" "mysql" {
  identifier             = "${var.service_name}-db"
  engine                 = "mysql"
  engine_version         = "8.0"
  instance_class         = "db.t3.micro" # Free tier eligible
  allocated_storage      = 20
  db_name                = var.db_name
  username               = var.db_username
  password               = var.db_password
  db_subnet_group_name   = aws_db_subnet_group.rds.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false

  # Settings for ephemeral/dev environments
  skip_final_snapshot    = true
  deletion_protection    = false

  tags = {
    Name = "${var.service_name}-db"
  }
}
