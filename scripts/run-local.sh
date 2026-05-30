#!/bin/bash

# run-local.sh - Spins up the full RTSCS stack locally
# Sets up Kind cluster and deploys all services

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

echo "RTSCS Local Development Setup"
echo "============================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Step 1: Create Kind cluster
echo -e "${YELLOW}Step 1: Creating Kind cluster...${NC}"
if kind get clusters | grep -q "rtscs-local"; then
    echo "Cluster 'rtscs-local' already exists, skipping creation"
else
    kind create cluster --name rtscs-local --config - <<EOF
apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
name: rtscs-local
nodes:
  - role: control-plane
    image: kindest/node:v1.29.2
    extraPortMappings:
      - containerPort: 80
        hostPort: 80
      - containerPort: 443
        hostPort: 443
      - containerPort: 8080
        hostPort: 8080
  - role: worker
    image: kindest/node:v1.29.2
  - role: worker
    image: kindest/node:v1.29.2
EOF
fi
echo -e "${GREEN}✓ Kind cluster ready${NC}"
echo ""

# Step 2: Switch context
echo -e "${YELLOW}Step 2: Switching kubectl context...${NC}"
kubectl cluster-info --context kind-rtscs-local
echo -e "${GREEN}✓ Context switched${NC}"
echo ""

# Step 3: Apply system manifests
echo -e "${YELLOW}Step 3: Applying system manifests...${NC}"
kubectl apply -f infra/k8s/system/
echo -e "${GREEN}✓ System manifests applied${NC}"
echo ""

# Step 4: Initialize Terraform (optional, for infrastructure)
echo -e "${YELLOW}Step 4: Initializing Terraform...${NC}"
cd infra/terraform
if [ ! -d ".terraform" ]; then
    terraform init
fi
cd "$REPO_ROOT"
echo -e "${GREEN}✓ Terraform initialized${NC}"
echo ""

# Step 5: Deploy with Terraform
echo -e "${YELLOW}Step 5: Deploying infrastructure with Terraform...${NC}"
cd infra/terraform
# Note: This is a simplified version. For full automation, uncomment:
# terraform apply -auto-approve -var-file=local.tfvars
echo "To deploy infrastructure, run:"
echo "  cd infra/terraform"
echo "  terraform apply -var-file=local.tfvars"
cd "$REPO_ROOT"
echo ""

# Step 6: Port forwarding instructions
echo -e "${YELLOW}Step 6: Setting up port forwarding...${NC}"
echo "Open a new terminal and run these commands:"
echo ""
echo "# Argo CD (port 8080)"
echo "kubectl port-forward -n argocd svc/argocd-server 8080:443"
echo ""
echo "# Grafana (port 3000)"
echo "kubectl port-forward -n observability svc/grafana 3000:80"
echo ""
echo "# Jaeger (port 16686)"
echo "kubectl port-forward -n observability svc/jaeger-query 16686:16686"
echo ""
echo "# API Gateway (port 8000)"
echo "kubectl port-forward -n rtscs-services svc/api-gateway 8000:8000"
echo ""

# Step 7: Verification
echo -e "${YELLOW}Step 7: Verifying cluster...${NC}"
echo "Cluster info:"
kubectl cluster-info --context kind-rtscs-local
echo ""
echo "Namespaces:"
kubectl get ns
echo ""

echo -e "${GREEN}===== Setup Complete =====${NC}"
echo ""
echo "Access points:"
echo "  - API Gateway: http://localhost:8000"
echo "  - Frontend: http://localhost:3000 (after deployment)"
echo "  - Grafana: http://localhost:3000 (port-forward needed)"
echo "  - Jaeger: http://localhost:16686 (port-forward needed)"
echo "  - Argo CD: http://localhost:8080 (port-forward needed)"
echo ""
echo "To view logs:"
echo "  kubectl logs -f deployment/inventory-service -n rtscs-services"
echo ""
echo "To clean up:"
echo "  kind delete cluster --name rtscs-local"
echo ""
