variable "namespace" {
  type = string
}

variable "postgres_password" {
  type      = string
  sensitive = true
  default   = "postgres"
}

variable "storage_class" {
  type    = string
  default = "standard"
}

variable "storage_size" {
  type    = string
  default = "10Gi"
}
