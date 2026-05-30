resource "helm_release" "argocd" {
  name             = "argocd"
  repository       = "https://argoproj.github.io/argo-helm"
  chart            = "argo-cd"
  version          = var.argocd_version
  namespace        = var.namespace
  create_namespace = true

  set {
    name  = "configs.secret.argocdServerAdminPassword"
    value = bcrypt("admin")  # Default admin password
  }

  set {
    name  = "server.insecure"
    value = "true"
  }

  set {
    name  = "server.ingress.enabled"
    value = "false"  # Use port-forward instead for local dev
  }

  # Enable server persistence
  set {
    name  = "server.statefulset.enabled"
    value = "true"
  }
}

# Create AppProject for RTSCS services
resource "kubectl_manifest" "rtscs_project" {
  yaml_body = <<-YAML
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: rtscs
  namespace: ${var.namespace}
spec:
  sourceRepos:
  - '${var.github_repo_url}'
  destinations:
  - namespace: 'rtscs-*'
    server: https://kubernetes.default.svc
  - namespace: observability
    server: https://kubernetes.default.svc
  - namespace: argocd
    server: https://kubernetes.default.svc
YAML

  depends_on = [helm_release.argocd]
}

# Create root Application to auto-sync all services
resource "kubectl_manifest" "rtscs_root_app" {
  yaml_body = <<-YAML
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: rtscs
  namespace: ${var.namespace}
spec:
  project: rtscs
  source:
    repoURL: ${var.github_repo_url}
    targetRevision: main
    path: infra/k8s/apps
  destination:
    server: https://kubernetes.default.svc
    namespace: rtscs-services
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
YAML

  depends_on = [kubectl_manifest.rtscs_project]
}
