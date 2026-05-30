variable "cluster_name" {
  description = "Kind cluster name"
  type        = string
  default     = "rtscs-local"
}

variable "cluster_context" {
  description = "Kubernetes context name"
  type        = string
  default     = "kind-rtscs-local"
}

variable "kubeconfig_path" {
  description = "Path to kubeconfig file"
  type        = string
  default     = "~/.kube/config"
}

variable "kind_image_version" {
  description = "Kind node image version"
  type        = string
  default     = "kindest/node:v1.29.2"
}
