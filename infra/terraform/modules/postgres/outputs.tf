output "host" {
  value = "postgres.${var.namespace}.svc.cluster.local"
}

output "port" {
  value = 5432
}

output "username" {
  value = "postgres"
}
