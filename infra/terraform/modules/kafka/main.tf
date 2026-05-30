# Install Strimzi Kafka operator
resource "helm_release" "kafka_operator" {
  name             = "strimzi"
  repository       = "https://strimzi.io/charts"
  chart            = "strimzi-kafka-operator"
  version          = "0.40.0"
  namespace        = var.namespace
  create_namespace = true
}

# Create Kafka cluster
resource "kubectl_manifest" "kafka_cluster" {
  yaml_body = <<-YAML
apiVersion: kafka.strimzi.io/v1beta2
kind: Kafka
metadata:
  name: kafka
  namespace: ${var.namespace}
spec:
  kafka:
    version: 3.6.0
    replicas: ${var.kafka_replicas}
    listeners:
      - name: plain
        port: 9092
        type: internal
        tls: false
    config:
      auto.create.topics.enable: "true"
      log.retention.hours: 720
      log.retention.bytes: 1073741824
    storage:
      type: persistent-claim
      size: ${var.broker_storage_size}
      class: standard
  zookeeper:
    replicas: 3
    storage:
      type: persistent-claim
      size: ${var.zookeeper_storage_size}
      class: standard
  entityOperator:
    topicOperator: {}
    userOperator: {}
YAML

  depends_on = [helm_release.kafka_operator]
}
