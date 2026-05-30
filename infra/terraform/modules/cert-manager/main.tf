# Create namespace for cert-manager
resource "kubernetes_namespace" "cert_manager" {
  metadata {
    name = var.namespace
    labels = {
      name = var.namespace
    }
  }
}

# Install cert-manager via Helm
resource "helm_release" "cert_manager" {
  name             = "cert-manager"
  repository       = "https://charts.jetstack.io"
  chart            = "cert-manager"
  version          = var.cert_manager_version
  namespace        = kubernetes_namespace.cert_manager.metadata[0].name
  create_namespace = true

  set {
    name  = "installCRDs"
    value = "true"
  }

  set {
    name  = "global.leaderElection.namespace"
    value = kubernetes_namespace.cert_manager.metadata[0].name
  }

  depends_on = [kubernetes_namespace.cert_manager]
}
