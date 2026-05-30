# RTSCS: Real-Time Supply Chain System

A modern, scalable, polyglot supply chain platform built with **gRPC** (internal synchronous communication), **Kafka** (event-driven async), **Kubernetes** (orchestration), and a **Next.js** frontend with real-time updates via WebSocket.

## Architecture Overview

```
[Users / Partners] 
       ↓ REST/JSON
[API Gateway (Bun)]
       ↓ gRPC (internal)
       ├─ Inventory Service (Java)
       ├─ Order Service (Java)
       ├─ Logistics Service (Rust)
       ├─ Notification Service (Go)
       └─ Analytics Service (Python)
       
       ↓ Kafka (async events)
       ├─ Notification Service (listen for order/shipment events)
       ├─ Analytics Service (aggregate metrics)
       └─ [Frontend WebSocket] via Redis Pub/Sub
       
[Frontend: Next.js] ← WebSocket (real-time updates)
```

**Key Design Decisions**:
- **gRPC inside the cluster** for low-latency, typed, polyglot sync communication
- **REST/JSON** exposed only at the edge via API Gateway
- **Kafka** for event-driven, eventual consistency, and real-time flows
- **Kubernetes + Argo CD** for GitOps automation
- **Buf** for protobuf schema management with breaking change detection
- **Observability** via OpenTelemetry + Jaeger, Prometheus, Grafana, Loki

## Quick Start

### Prerequisites
- Docker Desktop (or Docker + Docker Compose)
- Terraform 1.5+
- kubectl
- Bun or Node.js 20+
- Go 1.22+, Java 21+, Rust 1.75+, Python 3.11+

### Setup Local Development Environment

```bash
# 1. Clone and navigate to repo
git clone https://github.com/yourorg/rtscs
cd rtscs

# 2. Install tools (Buf, protoc, Kind, Helm, etc.)
bash scripts/install-tools.sh

# 3. Generate code from .proto files
bash scripts/gen-proto.sh

# 4. Spin up Kind cluster with all dependencies
bash scripts/run-local.sh

# 5. Access services
# - API Gateway: http://localhost:8000
# - Frontend: http://localhost:3000
# - Keycloak: http://localhost:8080
# - Grafana: http://localhost:3000 (admin/admin)
# - Jaeger: http://localhost:16686
```

## Project Structure

- **`proto/`** – Shared Protobuf definitions (.proto files)
  - `v1/common/` – Shared messages (User, Error, Timestamp)
  - `v1/inventory/`, `v1/order/`, `v1/logistics/`, etc. – Service APIs
  - `v1/events/` – Kafka event schemas
- **`services/`** – Microservices (Java, Rust, Go, Python)
  - `inventory-service/` – Stock management (Java)
  - `order-service/` – Order orchestration (Java)
  - `logistics-service/` – Shipment tracking (Rust)
  - `notification-service/` – Email/SMS/Push (Go)
  - `analytics-service/` – Metrics & forecasting (Python)
  - `api-gateway/` – REST proxy + WebSocket gateway (Bun)
- **`frontend/`** – Next.js React application
- **`infra/`** – Infrastructure as Code
  - `terraform/` – Terraform modules (Kind, cert-manager, Argo CD, databases)
  - `k8s/` – Kubernetes manifests (apps, system, observability)
  - `scripts/` – Setup automation
- **`docs/`** – Architecture, setup, API, debugging guides
- **`scripts/`** – Development tooling (Makefile, gen-proto, install-tools)

## Development Workflow

### Working with Protobuf

```bash
# Define new messages/services in .proto files
nano proto/v1/my_service/messages.proto

# Validate syntax and check for breaking changes
buf lint proto/
buf breaking proto/ --against-input='.git#branch=main'

# Generate code for all languages
bash scripts/gen-proto.sh
# Auto-generates:
# - Java: services/*/src/main/java/gen/
# - Rust: services/logistics-service/src/gen/
# - Go: services/notification-service/gen/
# - Python: services/analytics-service/gen/
# - TypeScript: services/api-gateway/src/gen/
```

### Testing

```bash
# Run all tests
make test

# Run specific service tests
make test-inventory
make test-order
make test-logistics

# E2E tests (deployed to Kind)
make test-e2e
```

### Deployment

```bash
# All changes are auto-synced via Argo CD
# Push to main branch → GitHub Actions builds → Image pushed → Argo CD syncs

# Manual sync (if needed)
argocd app sync rtscs-services
```

## Documentation

- **[`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)** – Detailed design, rationale, data flow
- **[`docs/SETUP.md`](docs/SETUP.md)** – Local development environment setup
- **[`docs/PROTO_GUIDE.md`](docs/PROTO_GUIDE.md)** – Protobuf workflow, regenerating code
- **[`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md)** – GitOps workflow, Argo CD
- **[`docs/API_GUIDE.md`](docs/API_GUIDE.md)** – REST endpoints, gRPC services
- **[`docs/GRPC_DEBUGGING.md`](docs/GRPC_DEBUGGING.md)** – Testing with `grpcurl`, debugging
- **[`docs/OBSERVABILITY.md`](docs/OBSERVABILITY.md)** – Jaeger, Prometheus, Grafana, Loki
- **[`docs/SECURITY.md`](docs/SECURITY.md)** – mTLS, RBAC, network policies, secrets

## Technology Stack

| Layer | Technology |
|-------|-----------|
| **Frontend** | Next.js 14, React 18, TypeScript, Tailwind CSS, TanStack Query |
| **API Gateway** | Bun, TypeScript, gRPC clients |
| **Services** | Java (Spring Boot), Rust (Tonic), Go (stdlib), Python (FastAPI) |
| **Communication** | gRPC (sync), Kafka (async), WebSocket (real-time) |
| **Databases** | PostgreSQL (transactional), Redis (cache + pub/sub), MongoDB/Kafka (events) |
| **Orchestration** | Kubernetes (Kind locally, EKS/GKE/AKS in production) |
| **GitOps** | Argo CD |
| **Observability** | Jaeger (traces), Prometheus (metrics), Grafana (dashboards), Loki (logs) |
| **Protobuf** | Buf CLI |
| **CI/CD** | GitHub Actions |
| **IaC** | Terraform |

## Phase Timeline

| Phase | Duration | Focus | Status |
|-------|----------|-------|--------|
| **0** | Week 1–2 | Monorepo, .proto, Terraform, Argo CD | 🟡 In Progress |
| **1** | Week 3–7 | Inventory + Order services, gRPC, mTLS | 🔴 Pending |
| **2** | Week 8–11 | Logistics, Notification, Analytics, Kafka | 🔴 Pending |
| **3** | Week 12–16 | API Gateway, Next.js, WebSocket | 🔴 Pending |
| **4** | Week 17–20 | Istio, SLOs, Security, Observability | 🔴 Pending |
| **5** | Week 21–24 | Testing, Performance, Production Readiness | 🔴 Pending |

## Contributing

1. Create a feature branch: `git checkout -b feature/your-feature`
2. Define or update .proto files as needed
3. Run `buf lint` and `buf breaking` to validate
4. Implement changes in service(s)
5. Run tests: `make test`
6. Commit with a clear message
7. Push and create a Pull Request

See [`CONTRIBUTING.md`](docs/CONTRIBUTING.md) for detailed guidelines.

## Support & Feedback

- **Issues**: Open a GitHub issue for bugs or feature requests
- **Discussions**: Use GitHub Discussions for architecture questions
- **Slack**: Join #rtscs-dev (if using workspace)

## License

[Your License Here] – See LICENSE file

---

**Last Updated**: May 30, 2026  
**Maintained by**: [Your Team]
