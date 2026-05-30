resource "helm_release" "redis" {
  name             = "redis"
  repository       = "https://charts.bitnami.com/bitnami"
  chart            = "redis"
  version          = "18.0.0"
  namespace        = var.namespace
  create_namespace = true

  set {
    name  = "architecture"
    value = "standalone"
  }

  set {
    name  = "auth.enabled"
    value = "false"
  }

  set {
    name  = "master.persistence.enabled"
    value = "true"
  }

  set {
    name  = "master.persistence.storageClass"
    value = var.storage_class
  }

  set {
    name  = "master.persistence.size"
    value = var.storage_size
  }

  set {
    name  = "metrics.enabled"
    value = "true"
  }
}
