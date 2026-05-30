# RTSCS Local Development Setup

This guide walks through setting up RTSCS for local development using Kind (Kubernetes in Docker).

## Prerequisites

Before starting, ensure you have:

- **Docker Desktop** (or Docker + Docker Compose) — https://docker.com/products/docker-desktop
- **Homebrew** (macOS) or apt/yum (Linux)
- **Git** for version control
- **4GB+ RAM** available for Kind cluster (8GB recommended)
- **10GB+ free disk space**

## Installation (One-Time Setup)

### 1. Install Development Tools

Run the automated installer:

```bash
cd rtscs
bash scripts/install-tools.sh
```

This installs:
- kubectl
- Kind
- Helm
- Terraform
- Buf (Protobuf toolchain)
- protoc
- Language tools (Java, Go, Rust, Python, Bun)

**Or, install manually:**

#### macOS (using Homebrew)
```bash
brew install kubectl kind helm terraform protobuf buf
brew install go rust python@3.11
# Java
brew install openjdk@21

# Bun
curl -fsSL https://bun.sh/install | bash
export PATH="$HOME/.bun/bin:$PATH"
```

#### Linux (Ubuntu/Debian)
```bash
# Docker
curl -fsSL https://get.docker.com -o get-docker.sh
bash get-docker.sh

# Tools
sudo apt-get install -y curl wget git
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Kind, Helm, Terraform, etc. (see scripts/install-tools.sh for commands)
```

### 2. Clone Repository

```bash
git clone https://github.com/yourorg/rtscs
cd rtscs
```

### 3. Generate Protobuf Code

```bash
bash scripts/gen-proto.sh
```

This generates code for Java, Go, Rust, Python, and TypeScript from `.proto` files.

## Running Locally

### Quick Start (All-in-One)

```bash
make run-local
```

This script:
1. Creates a Kind cluster named `rtscs-local`
2. Applies Kubernetes system manifests
3. Initializes Terraform
4. Provides port-forward instructions

### Or, Step-by-Step

#### Step 1: Create Kind Cluster

```bash
make setup-kind
```

Verify:
```bash
kubectl cluster-info
kind get clusters  # Should show: rtscs-local
```

#### Step 2: Check Kubectl Context

```bash
kubectl config current-context  # Should be: kind-rtscs-local
```

Switch if needed:
```bash
kubectl config use-context kind-rtscs-local
```

#### Step 3: Create Namespaces

```bash
kubectl apply -f infra/k8s/system/namespaces.yaml
kubectl get ns
```

#### Step 4: Deploy Infrastructure (Optional)

To deploy PostgreSQL, Redis, Kafka, cert-manager, and Argo CD via Terraform:

```bash
cd infra/terraform
terraform init
terraform apply -var-file=local.tfvars
```

Expected output:
- `rtscs_data` namespace with PostgreSQL, Redis, Kafka
- `argocd` namespace with Argo CD
- Certificates created for mTLS

**Without Terraform**, you can deploy individual services:
```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add strimzi https://strimzi.io/charts
helm repo add jetstack https://charts.jetstack.io

helm install postgres bitnami/postgresql -n rtscs-data --create-namespace
helm install redis bitnami/redis -n rtscs-data
helm install kafka strimzi/strimzi-kafka-operator -n rtscs-data
helm install cert-manager jetstack/cert-manager -n cert-manager --create-namespace --set installCRDs=true
```

#### Step 5: Apply System Manifests

```bash
kubectl apply -f infra/k8s/system/
```

This creates:
- Namespaces
- Network policies
- RBAC (ServiceAccounts, Roles)

#### Step 6: Deploy Services (Placeholder)

```bash
kubectl apply -k infra/k8s/apps/
```

Currently applies placeholder deployments. Real services deployed in Phase 1.

### Step 7: Port Forwarding

Open separate terminal windows:

```bash
# API Gateway (main entry point)
kubectl port-forward -n rtscs-services svc/api-gateway 8000:8000

# Argo CD (if deployed)
kubectl port-forward -n argocd svc/argocd-server 8080:443

# PostgreSQL (if deployed)
kubectl port-forward -n rtscs-data svc/postgres 5432:5432

# Redis (if deployed)
kubectl port-forward -n rtscs-data svc/redis-master 6379:6379

# Kafka (if deployed)
kubectl port-forward -n rtscs-data svc/kafka-broker 9092:9092

# Grafana (once deployed)
kubectl port-forward -n observability svc/grafana 3000:80

# Jaeger (once deployed)
kubectl port-forward -n observability svc/jaeger-query 16686:16686
```

### Step 8: Verify Cluster Status

```bash
make status
```

Or manually:
```bash
kubectl get nodes
kubectl get pods -A
kubectl get svc -A
```

## Common Commands

### Building Services (Phase 1+)

```bash
# Build all
make build-all

# Build individual service
make build-inventory
make build-order
make build-logistics
make build-notification
make build-analytics
make build-gateway
```

### Testing

```bash
# Test all
make test

# Test individual
make test-inventory
make test-logistics
```

### Viewing Logs

```bash
# View real-time logs
make logs-gateway

# Or manually
kubectl logs -f deployment/api-gateway -n rtscs-services

# Multiple pods
kubectl logs -l app=order-service -n rtscs-services
```

### Protobuf Development

```bash
# Edit .proto files in proto/v1/
nano proto/v1/order/service.proto

# Validate and generate code
make proto-lint
bash scripts/gen-proto.sh

# Check for breaking changes
make proto-breaking
```

### Deploying Updates

```bash
# After code changes
make build-inventory
docker build -t inventory-service:latest services/inventory-service/

# Update K8s
kubectl set image deployment/inventory-service \
  inventory-service=inventory-service:latest \
  -n rtscs-services

# Or via GitOps (Argo CD)
git add .
git commit -m "Update inventory service"
git push  # Argo CD auto-syncs
```

## Troubleshooting

### Kind Cluster Won't Start

```bash
# Check Docker daemon
docker ps

# Delete and recreate cluster
kind delete cluster --name rtscs-local
make setup-kind
```

### Pods in CrashLoopBackOff

```bash
# Check logs
kubectl logs <pod-name> -n rtscs-services
kubectl describe pod <pod-name> -n rtscs-services

# Common issues:
# - Missing environment variables (check Deployment env section)
# - Missing database (PostgreSQL not ready)
# - Image pull errors (check image name)
```

### Port 8000 Already in Use

```bash
# Find process
lsof -i :8000  # macOS/Linux
netstat -ano | findstr :8000  # Windows

# Kill process or use different port
kubectl port-forward -n rtscs-services svc/api-gateway 8001:8000
```

### Terraform Apply Fails

```bash
# Ensure kubeconfig is set
export KUBECONFIG=$HOME/.kube/config

# Try again
cd infra/terraform
terraform init
terraform apply -var-file=local.tfvars

# Debug
terraform plan  # Preview changes
terraform destroy  # Clean up if needed
```

### Out of Storage

```bash
# Check disk usage
docker system df

# Clean up unused images/volumes
docker system prune -a

# Delete cluster and rebuild
kind delete cluster --name rtscs-local
docker volume prune
```

## Environment Variables

Create `.env.local` for development:

```bash
# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres123

# Kafka
KAFKA_BROKERS=localhost:9092

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# gRPC Services
INVENTORY_SERVICE_URL=localhost:50051
ORDER_SERVICE_URL=localhost:50051
LOGISTICS_SERVICE_URL=localhost:50051

# API Gateway
API_GATEWAY_PORT=8000
NODE_ENV=development
LOG_LEVEL=INFO
```

## Next Steps

1. **Generate Proto Code**: `bash scripts/gen-proto.sh`
2. **Start Building Services**: See Phase 1 tasks
3. **Read Architecture Docs**: `docs/ARCHITECTURE.md`
4. **Explore Debugging**: `docs/GRPC_DEBUGGING.md`

## Getting Help

- Check `/docs/` for comprehensive guides
- View service-specific READMEs in `/services/*/README.md`
- Run `make help` for all available commands
- Open GitHub issues for bugs or questions

---

**Document Version**: 1.0  
**Last Updated**: May 30, 2026
