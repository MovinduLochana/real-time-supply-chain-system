output "namespace" {
  value = var.namespace
}

output "server_service_name" {
  value = "argocd-server"
}

output "access_instructions" {
  value = "Run: kubectl port-forward -n ${var.namespace} svc/argocd-server 8080:443"
}
