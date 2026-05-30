variable "namespace" {
  description = "Namespace where cert-manager will be installed"
  type        = string
  default     = "cert-manager"
}

variable "cert_manager_version" {
  description = "cert-manager version"
  type        = string
  default     = "v1.13.2"
}
