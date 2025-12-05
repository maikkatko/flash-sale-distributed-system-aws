output "order_queue_arn" {
  description = "The ARN of the main order queue."
  value       = aws_sqs_queue.main.arn
}

output "order_queue_url" {
  description = "The URL of the main order queue."
  value       = aws_sqs_queue.main.id # .id returns the URL for SQS queues
}

output "dlq_arn" {
  description = "The ARN of the dead-letter queue."
  value       = aws_sqs_queue.dlq.arn
}

output "dlq_url" {
  description = "The URL of the dead-letter queue."
  value       = aws_sqs_queue.dlq.id
}
