output "cluster_name" {
  description = "Kind cluster name"
  value       = var.cluster_name
}

output "context_name" {
  description = "Kubernetes context name"
  value       = var.cluster_context
}

output "kubeconfig_path" {
  description = "Path to kubeconfig file"
  value       = var.kubeconfig_path
}
