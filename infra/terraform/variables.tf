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
  description = "Kind image version (e.g., kindest/node:v1.29.0)"
  type        = string
  default     = "kindest/node:v1.29.2"
}

variable "cert_manager_version" {
  description = "cert-manager Helm chart version"
  type        = string
  default     = "v1.13.2"
}

variable "postgres_password" {
  description = "PostgreSQL root password"
  type        = string
  sensitive   = true
  default     = "postgres"
}

variable "postgres_storage_size" {
  description = "PostgreSQL PVC size"
  type        = string
  default     = "10Gi"
}

variable "redis_storage_size" {
  description = "Redis PVC size"
  type        = string
  default     = "5Gi"
}

variable "kafka_replicas" {
  description = "Kafka broker replicas"
  type        = number
  default     = 3
}

variable "kafka_broker_storage_size" {
  description = "Kafka broker storage size"
  type        = string
  default     = "10Gi"
}

variable "kafka_zookeeper_storage_size" {
  description = "Kafka Zookeeper storage size"
  type        = string
  default     = "5Gi"
}

variable "argocd_version" {
  description = "Argo CD Helm chart version"
  type        = string
  default     = "6.0.0"
}

variable "argocd_domain" {
  description = "Argo CD domain (local development)"
  type        = string
  default     = "argocd.localhost"
}

variable "github_repo_url" {
  description = "GitHub repository URL for Argo CD"
  type        = string
  default     = "https://github.com/yourorg/rtscs"
}
