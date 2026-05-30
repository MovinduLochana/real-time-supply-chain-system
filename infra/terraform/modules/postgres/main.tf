resource "helm_release" "postgres" {
  name             = "postgres"
  repository       = "https://charts.bitnami.com/bitnami"
  chart            = "postgresql"
  version          = "13.2.0"
  namespace        = var.namespace
  create_namespace = true

  set {
    name  = "auth.postgresPassword"
    value = var.postgres_password
  }

  set {
    name  = "primary.persistence.storageClass"
    value = var.storage_class
  }

  set {
    name  = "primary.persistence.size"
    value = var.storage_size
  }

  set {
    name  = "primary.persistence.enabled"
    value = "true"
  }

  set {
    name  = "metrics.enabled"
    value = "true"
  }
}
