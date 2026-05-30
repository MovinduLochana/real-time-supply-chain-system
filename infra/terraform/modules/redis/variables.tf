variable "namespace" {
  type = string
}

variable "storage_class" {
  type    = string
  default = "standard"
}

variable "storage_size" {
  type    = string
  default = "5Gi"
}
