.PHONY: help install setup-kind proto-gen build test deploy-local clean

# Default target
.DEFAULT_GOAL := help

# Colors for output
CYAN := \033[0;36m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(CYAN)RTSCS Development Commands$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(CYAN)Examples:$(NC)"
	@echo "  make install         # Install all dependencies"
	@echo "  make setup-kind      # Create Kind cluster"
	@echo "  make proto-gen       # Generate code from .proto files"
	@echo "  make build-all       # Build all services"
	@echo "  make test            # Run all tests"
	@echo "  make deploy-local    # Deploy to local Kind cluster"
	@echo ""

install: ## Install development tools and dependencies
	@echo "$(YELLOW)Installing development tools...$(NC)"
	bash scripts/install-tools.sh

setup-kind: ## Create local Kind cluster
	@echo "$(YELLOW)Setting up Kind cluster...$(NC)"
	kind create cluster --name rtscs-local --config infra/kind-config.yaml 2>/dev/null || echo "Cluster already exists"
	kubectl cluster-info --context kind-rtscs-local
	@echo "$(GREEN)✓ Kind cluster ready$(NC)"

proto-gen: ## Generate code from .proto files
	@echo "$(YELLOW)Generating protobuf code...$(NC)"
	bash scripts/gen-proto.sh
	@echo "$(GREEN)✓ Code generation complete$(NC)"

proto-lint: ## Lint .proto files with Buf
	@echo "$(YELLOW)Linting .proto files...$(NC)"
	buf lint proto/

proto-breaking: ## Check for breaking changes in .proto files
	@echo "$(YELLOW)Checking for breaking changes...$(NC)"
	buf breaking proto/ --against-input='.git#branch=main'

build-all: ## Build all services
	@echo "$(YELLOW)Building all services...$(NC)"
	@echo "Note: Actual builds will be in Phase 1"
	@echo "$(GREEN)✓ Build targets configured$(NC)"

build-inventory: ## Build Inventory Service (Java)
	@echo "$(YELLOW)Building Inventory Service...$(NC)"
	cd services/inventory-service && gradle build

build-order: ## Build Order Service (Java)
	@echo "$(YELLOW)Building Order Service...$(NC)"
	cd services/order-service && gradle build

build-logistics: ## Build Logistics Service (Rust)
	@echo "$(YELLOW)Building Logistics Service...$(NC)"
	cd services/logistics-service && cargo build --release

build-notification: ## Build Notification Service (Go)
	@echo "$(YELLOW)Building Notification Service...$(NC)"
	cd services/notification-service && go build -o notification-service .

build-analytics: ## Build Analytics Service (Python)
	@echo "$(YELLOW)Building Analytics Service...$(NC)"
	cd services/analytics-service && pip install -r requirements.txt

build-gateway: ## Build API Gateway (Bun)
	@echo "$(YELLOW)Building API Gateway...$(NC)"
	cd services/api-gateway && bun install && bun run build

test: ## Run all tests
	@echo "$(YELLOW)Running tests...$(NC)"
	@echo "$(RED)Note: Tests will be implemented in Phase 1$(NC)"

test-inventory: ## Test Inventory Service
	@echo "$(YELLOW)Testing Inventory Service...$(NC)"
	cd services/inventory-service && gradle test

test-order: ## Test Order Service
	@echo "$(YELLOW)Testing Order Service...$(NC)"
	cd services/order-service && gradle test

test-logistics: ## Test Logistics Service
	@echo "$(YELLOW)Testing Logistics Service...$(NC)"
	cd services/logistics-service && cargo test

test-notification: ## Test Notification Service
	@echo "$(YELLOW)Testing Notification Service...$(NC)"
	cd services/notification-service && go test ./...

test-analytics: ## Test Analytics Service
	@echo "$(YELLOW)Testing Analytics Service...$(NC)"
	cd services/analytics-service && pytest tests/

test-gateway: ## Test API Gateway
	@echo "$(YELLOW)Testing API Gateway...$(NC)"
	cd services/api-gateway && bun test

test-e2e: ## Run end-to-end tests
	@echo "$(YELLOW)Running E2E tests...$(NC)"
	@echo "$(RED)Note: E2E tests will be implemented in Phase 5$(NC)"

deploy-local: ## Deploy to local Kind cluster
	@echo "$(YELLOW)Deploying to local Kind cluster...$(NC}"
	kubectl apply -f infra/k8s/system/
	kubectl apply -k infra/k8s/apps/
	@echo "$(GREEN)✓ Deployment complete$(NC)"
	@echo ""
	@echo "Port forwarding commands:"
	@echo "  kubectl port-forward -n argocd svc/argocd-server 8080:443"
	@echo "  kubectl port-forward -n rtscs-services svc/api-gateway 8000:8000"

deploy-infra: ## Deploy infrastructure with Terraform
	@echo "$(YELLOW)Deploying infrastructure...$(NC)"
	cd infra/terraform && terraform init && terraform apply -var-file=local.tfvars
	@echo "$(GREEN)✓ Infrastructure deployed$(NC)"

logs-inventory: ## View Inventory Service logs
	kubectl logs -f deployment/inventory-service -n rtscs-services

logs-order: ## View Order Service logs
	kubectl logs -f deployment/order-service -n rtscs-services

logs-logistics: ## View Logistics Service logs
	kubectl logs -f deployment/logistics-service -n rtscs-services

logs-notification: ## View Notification Service logs
	kubectl logs -f deployment/notification-service -n rtscs-services

logs-analytics: ## View Analytics Service logs
	kubectl logs -f deployment/analytics-service -n rtscs-services

logs-gateway: ## View API Gateway logs
	kubectl logs -f deployment/api-gateway -n rtscs-services

status: ## Check cluster and service status
	@echo "$(CYAN)Cluster Status:$(NC)"
	@kubectl cluster-info --context kind-rtscs-local 2>/dev/null || echo "Cluster not found. Run: make setup-kind"
	@echo ""
	@echo "$(CYAN)Namespaces:$(NC)"
	@kubectl get ns
	@echo ""
	@echo "$(CYAN)Services:$(NC)"
	@kubectl get svc -A
	@echo ""
	@echo "$(CYAN)Deployments:$(NC)"
	@kubectl get deployments -n rtscs-services

describe: ## Describe cluster resources
	@echo "$(CYAN)Nodes:$(NC)"
	@kubectl get nodes -o wide
	@echo ""
	@echo "$(CYAN)Pods:$(NC)"
	@kubectl get pods -A

clean: ## Clean up local Kind cluster and generated files
	@echo "$(RED)Cleaning up...$(NC)"
	kind delete cluster --name rtscs-local
	rm -rf proto/generated/
	rm -rf services/*/build
	rm -rf services/*/dist
	rm -rf services/*/target
	@echo "$(GREEN)✓ Cleanup complete$(NC)"

format: ## Format code
	@echo "$(YELLOW)Formatting code...$(NC}"
	@echo "$(RED)Note: Format targets will be configured per language in Phase 1$(NC)"

lint: ## Lint all code
	@echo "$(YELLOW)Linting code...$(NC}"
	buf lint proto/
	@echo "$(GREEN)✓ Proto files valid$(NC)"

docker-build-all: ## Build all Docker images locally
	@echo "$(YELLOW)Building Docker images...$(NC)"
	docker build -t rtscs/inventory-service:latest services/inventory-service/
	docker build -t rtscs/order-service:latest services/order-service/
	docker build -t rtscs/logistics-service:latest services/logistics-service/
	docker build -t rtscs/notification-service:latest services/notification-service/
	docker build -t rtscs/analytics-service:latest services/analytics-service/
	docker build -t rtscs/api-gateway:latest services/api-gateway/
	@echo "$(GREEN)✓ All Docker images built$(NC)"

docs: ## Open documentation in browser
	@echo "Opening documentation..."
	@[ -f docs/ARCHITECTURE.md ] && echo "Architecture docs available at docs/ARCHITECTURE.md"
	@[ -f docs/SETUP.md ] && echo "Setup guide available at docs/SETUP.md"

version: ## Show RTSCS version
	@echo "RTSCS v0.1.0-alpha (Phase 0: Foundations)"

.PHONY: proto-lint proto-breaking logs-inventory logs-order logs-logistics logs-notification logs-analytics logs-gateway
