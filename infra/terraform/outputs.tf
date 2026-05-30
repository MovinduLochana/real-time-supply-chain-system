output "kind_cluster_name" {
  description = "Kind cluster name"
  value       = module.k8s_cluster.cluster_name
}

output "kubeconfig_path" {
  description = "Path to kubeconfig file"
  value       = var.kubeconfig_path
}

output "argocd_password" {
  description = "Argo CD admin password (run: kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d)"
  value       = "Get via: kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d"
}

output "postgres_host" {
  description = "PostgreSQL host"
  value       = "postgres.rtscs-data.svc.cluster.local"
}

output "postgres_port" {
  description = "PostgreSQL port"
  value       = 5432
}

output "redis_host" {
  description = "Redis host"
  value       = "redis.rtscs-data.svc.cluster.local"
}

output "redis_port" {
  description = "Redis port"
  value       = 6379
}

output "kafka_broker_urls" {
  description = "Kafka broker connection strings"
  value       = "kafka.rtscs-data.svc.cluster.local:9092"
}

output "argocd_server_url" {
  description = "Argo CD server URL"
  value       = "http://localhost:8080 (port-forward: kubectl port-forward -n argocd svc/argocd-server 8080:443)"
}

output "grafana_url" {
  description = "Grafana URL"
  value       = "http://localhost:3000 (port-forward required)"
}

output "jaeger_url" {
  description = "Jaeger UI URL"
  value       = "http://localhost:16686 (port-forward required)"
}

output "environment_info" {
  description = "Environment configuration summary"
  value       = {
    cluster_name = module.k8s_cluster.cluster_name
    namespaces = [
      kubernetes_namespace.rtscs_services.metadata[0].name,
      kubernetes_namespace.rtscs_data.metadata[0].name,
      kubernetes_namespace.observability.metadata[0].name,
      kubernetes_namespace.argocd.metadata[0].name
    ]
    database_host = "postgres.rtscs-data.svc.cluster.local"
    redis_host    = "redis.rtscs-data.svc.cluster.local"
    kafka_broker  = "kafka.rtscs-data.svc.cluster.local:9092"
  }
}
