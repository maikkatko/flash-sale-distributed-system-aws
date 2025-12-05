resource "aws_sqs_queue" "dlq" {
  name                        = "${var.name_prefix}-dlq"
  message_retention_seconds   = var.message_retention_seconds
  tags                        = var.tags
}

resource "aws_sqs_queue" "main" {
  name              = var.name_prefix
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq.arn
    maxReceiveCount     = var.max_receive_count
  })

  # The order processor will poll this queue, so a longer receive wait time
  # reduces the number of empty receives and lowers costs.
  receive_wait_time_seconds = 20

  tags = merge(
    var.tags,
    {
      "IsDLQ" = "false"
    }
  )
}

