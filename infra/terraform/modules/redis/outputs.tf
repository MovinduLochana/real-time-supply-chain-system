output "host" {
  value = "redis.${var.namespace}.svc.cluster.local"
}

output "port" {
  value = 6379
}
