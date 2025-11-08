# /terraform/modules/rds/outputs.tf

output "db_endpoint" {
  description = "The connection endpoint for the RDS instance."
  value       = aws_db_instance.mysql.endpoint
}

output "db_name" {
  description = "The name of the database."
  value       = aws_db_instance.mysql.db_name
}

