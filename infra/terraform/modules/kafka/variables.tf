variable "namespace" {
  type = string
}

variable "kafka_replicas" {
  type    = number
  default = 3
}

variable "broker_storage_size" {
  type    = string
  default = "10Gi"
}

variable "zookeeper_storage_size" {
  type    = string
  default = "5Gi"
}
