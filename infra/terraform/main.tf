terraform {
  required_version = ">= 1.5"
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.25"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.12"
    }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = "~> 1.14"
    }
    null = {
      source  = "hashicorp/null"
      version = "~> 3.2"
    }
  }
}

provider "kubernetes" {
  config_path = var.kubeconfig_path
}

provider "helm" {
  kubernetes {
    config_path = var.kubeconfig_path
  }
}

provider "kubectl" {
  config_path            = var.kubeconfig_path
  load_config_file       = true
  apply_retry_count      = 5
}

# Create Kind cluster
module "k8s_cluster" {
  source = "./modules/k8s-cluster"

  cluster_name           = var.cluster_name
  cluster_context        = var.cluster_context
  kubeconfig_path        = var.kubeconfig_path
  kind_image_version     = var.kind_image_version
}

# Create namespaces
resource "kubernetes_namespace" "rtscs_services" {
  metadata {
    name = "rtscs-services"
    labels = {
      "name" = "rtscs-services"
    }
  }

  depends_on = [module.k8s_cluster]
}

resource "kubernetes_namespace" "rtscs_data" {
  metadata {
    name = "rtscs-data"
    labels = {
      "name" = "rtscs-data"
    }
  }

  depends_on = [module.k8s_cluster]
}

resource "kubernetes_namespace" "observability" {
  metadata {
    name = "observability"
    labels = {
      "name" = "observability"
    }
  }

  depends_on = [module.k8s_cluster]
}

resource "kubernetes_namespace" "argocd" {
  metadata {
    name = "argocd"
    labels = {
      "name" = "argocd"
    }
  }

  depends_on = [module.k8s_cluster]
}

# Install cert-manager
module "cert_manager" {
  source = "./modules/cert-manager"

  namespace              = "cert-manager"
  cert_manager_version   = var.cert_manager_version

  depends_on = [module.k8s_cluster]
}

# Install PostgreSQL
module "postgres" {
  source = "./modules/postgres"

  namespace           = kubernetes_namespace.rtscs_data.metadata[0].name
  postgres_password   = var.postgres_password
  storage_class       = "standard"
  storage_size        = var.postgres_storage_size

  depends_on = [kubernetes_namespace.rtscs_data, module.cert_manager]
}

# Install Redis
module "redis" {
  source = "./modules/redis"

  namespace     = kubernetes_namespace.rtscs_data.metadata[0].name
  storage_class = "standard"
  storage_size  = var.redis_storage_size

  depends_on = [kubernetes_namespace.rtscs_data, module.cert_manager]
}

# Install Kafka
module "kafka" {
  source = "./modules/kafka"

  namespace     = kubernetes_namespace.rtscs_data.metadata[0].name
  kafka_replicas = var.kafka_replicas
  broker_storage_size = var.kafka_broker_storage_size
  zookeeper_storage_size = var.kafka_zookeeper_storage_size

  depends_on = [kubernetes_namespace.rtscs_data, module.cert_manager]
}

# Install Argo CD
module "argo_cd" {
  source = "./modules/argo-cd"

  namespace           = kubernetes_namespace.argocd.metadata[0].name
  argocd_version      = var.argocd_version
  argocd_domain       = var.argocd_domain
  github_repo_url     = var.github_repo_url

  depends_on = [kubernetes_namespace.argocd, module.cert_manager]
}

# Create self-signed ClusterIssuer for local development
resource "kubectl_manifest" "self_signed_issuer" {
  yaml_body = <<-YAML
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
YAML

  depends_on = [module.cert_manager]
}

# Create storage class (local-path for Kind)
resource "kubectl_manifest" "local_storage_class" {
  yaml_body = <<-YAML
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-path
provisioner: rancher.io/local-path
allowVolumeExpansion: true
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
YAML

  depends_on = [module.k8s_cluster]
}
