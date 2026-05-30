# This module creates a Kind cluster configuration
# The actual cluster creation is done via a shell script (setup-kind.sh)
# This module mainly validates prerequisites and creates the Kind config file

resource "null_resource" "kind_cluster" {
  provisioners {
    local-exec {
      command = "echo 'Kind cluster module initialized. Run scripts/setup-kind.sh to create cluster'"
    }
  }
}

# Kind cluster config that will be used by setup-kind.sh
locals {
  kind_config = yamlencode({
    apiVersion = "kind.x-k8s.io/v1alpha4"
    kind       = "Cluster"
    name       = var.cluster_name
    nodes = [
      {
        role = "control-plane"
        image = var.kind_image_version
        extraPortMappings = [
          {
            containerPort = 80
            hostPort      = 80
            protocol      = "TCP"
          },
          {
            containerPort = 443
            hostPort      = 443
            protocol      = "TCP"
          },
          {
            containerPort = 8080
            hostPort      = 8080
            protocol      = "TCP"
          }
        ]
      },
      {
        role = "worker"
        image = var.kind_image_version
      },
      {
        role = "worker"
        image = var.kind_image_version
      }
    ]
    containerdConfigPatches = [
      <<-EOF
[plugins."io.containerd.grpc.v1.cri".registries.mirrors."localhost:5000"]
  endpoint = ["http://localhost:5000"]
EOF
    ]
  })
}

# Output the Kind config to a file for use in setup script
output "kind_config_yaml" {
  description = "Kind cluster configuration"
  value       = local.kind_config
}
