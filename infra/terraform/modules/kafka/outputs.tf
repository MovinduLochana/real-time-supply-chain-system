output "broker_urls" {
  value = "kafka-bootstrap.${var.namespace}.svc.cluster.local:9092"
}

output "bootstrap_servers" {
  value = "kafka-bootstrap.${var.namespace}.svc.cluster.local:9092"
}
