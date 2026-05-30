# Local development Terraform variables
cluster_name           = "rtscs-local"
cluster_context        = "kind-rtscs-local"
kubeconfig_path        = "~/.kube/config"
kind_image_version     = "kindest/node:v1.29.2"

cert_manager_version   = "v1.13.2"

postgres_password      = "postgres123"
postgres_storage_size  = "10Gi"

redis_storage_size     = "5Gi"

kafka_replicas         = 3
kafka_broker_storage_size = "10Gi"
kafka_zookeeper_storage_size = "5Gi"

argocd_version         = "6.0.0"
argocd_domain          = "argocd.localhost"
github_repo_url        = "https://github.com/yourorg/rtscs"
