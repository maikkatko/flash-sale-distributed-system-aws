variable "name_prefix" {
  description = "A prefix used to name the SQS queues."
  type        = string
}

variable "max_receive_count" {
  description = "The number of times a message is delivered to the source queue before being moved to the DLQ."
  type        = number
  default     = 5
}

variable "message_retention_seconds" {
  description = "The number of seconds to retain a message in the DLQ."
  type        = number
  default     = 1209600 # 14 days
}

variable "tags" {
  description = "A map of tags to assign to the resources."
  type        = map(string)
  default     = {}
}
