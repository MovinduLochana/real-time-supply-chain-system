# Real Time Supply Chain Management System

####  Project Description

Designed and developed a highly scalable, real-time **supply chain management platform** supporting **order tracking**, **inventory reservation**, **shipment coordination**, **notifications**, and **AI-driven forecasting**. Built using modern **microservices architecture** with full observability, event-driven processing, and secure authentication & authorization.

####  Key Features

-   Real-time order-to-delivery pipeline using Kafka
-   MicroServices with gRPC communication
-   GraphQL API Gateway for frontend integration
-   Role-based access control via OIDC (Keycloak)
-   Telemetry ingestion and ML-based forecasting
-   Fully containerized and orchestrated on Kubernetes

----------

###  Technologies Used

#### **Backend**

-   **Java 21** with Spring Boot including Security, Actuator and Cloud
-   **Kafka** for event-driven communication
-   **Redis** for caching and token/session store
-   **PostgreSQL** and **MongoDB** for structured and unstructured data
-   **Python (FastAPI / ML)** for anomaly detection and demand prediction
-   **Rust (Tokio/Axum)** for high-performance telemetry ingestion
-   **Go** for lightweight internal services (optional modules)

#### **Frontend**

-   Next.js (React) with bun 
-   Flutter App for mobile
-   Auth.js (NextAuth) integration with **Keycloak OIDC**
-   Fully responsive admin dashboard and mobile-first views
-   Real-time UI updates via GraphQL subscriptions and React Query

#### **Authentication / Authorization**

-   **Keycloak** for identity management (OIDC, OAuth2, RBAC)
-   Role-based route and API access across frontend and backend
    
#### **DevOps & Infrastructure**

-   **Docker**, **Helm**, **Kubernetes** for containerization and orchestration
-   **CI/CD** via **Jenkins**
-   **Terraform** & **Ansible** for infrastructure-as-code and provisioning
-   **Prometheus**, **Grafana** for observability
    
----------

###  Architecture Highlights

-   Microservice-based architecture with logical domain boundaries:
    -   Order, Inventory, Shipment, Notification, Telemetry, Forecasting
-   GraphQL Gateway aggregates service data for frontend
-   Service Discovery via Kubernetes DNS (no need for Eureka)
-   API Gateway (Spring Cloud Gateway) handles routing and security
-   Secure with JWT-based stateless authentication

---
Cuurently using REST and Euraka for easy testing, these will be replaced after working prototype is done


A comprehensive, polyglot microservices architecture for logistics and supply chain management.



\## Architecture Overview



```

┌─────────────────────────────────────────────────────────────────────────────────┐

│                              FRONTEND LAYER                                       │

│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │

│  │ Admin Portal │  │  Driver App  │  │ Warehouse App│  │ Customer App │        │

│  │  (Next.js)   │  │(React Native)│  │(React Native)│  │(React Native)│        │

│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘        │

└─────────┼─────────────────┼─────────────────┼─────────────────┼─────────────────┘

&nbsp;         │                 │                 │                 │

&nbsp;         └─────────────────┴────────┬────────┴─────────────────┘

&nbsp;                                    ▼

┌─────────────────────────────────────────────────────────────────────────────────┐

│                            API GATEWAY (Rust)                                    │

│                    JWT Validation │ Rate Limiting │ Routing                      │

└─────────────────────────────────────┬───────────────────────────────────────────┘

&nbsp;                                     │

&nbsp;         ┌───────────────────────────┼───────────────────────────┐

&nbsp;         ▼                           ▼                           ▼

┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐

│   Auth Service   │     │  Order Service   │     │Warehouse Service │

│     (Java)       │     │     (Java)       │     │     (Java)       │

│   + Keycloak     │     │   + PostgreSQL   │     │   + PostgreSQL   │

└──────────────────┘     └──────────────────┘     └──────────────────┘

&nbsp;         │                        │                        │

&nbsp;         ▼                        ▼                        ▼

┌─────────────────────────────────────────────────────────────────────────────────┐

│                              KAFKA EVENT BUS                                     │

└───────────────────────────────────┬─────────────────────────────────────────────┘

&nbsp;         ▲                         │                         ▲

&nbsp;         │                         ▼                         │

┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐

│  GPS Ingestion   │     │Analytics Consumer│     │ Report Generator │

│     (Rust)       │     │    (Python)      │     │    (Python)      │

│  High-Throughput │     │  + TimescaleDB   │     │   PDF/Excel      │

└──────────────────┘     └──────────────────┘     └──────────────────┘

&nbsp;                                 │

&nbsp;                                 ▼

&nbsp;                        ┌──────────────────┐

&nbsp;                        │   ML Pipelines   │

&nbsp;                        │    (Python)      │

&nbsp;                        │ Anomaly Detection│

&nbsp;                        └──────────────────┘



┌─────────────────────────────────────────────────────────────────────────────────┐

│                         CROSS-CUTTING SERVICES (Go)                              │

│         Config Service │ Health Service │ Metrics (Prometheus/Grafana)          │

└─────────────────────────────────────────────────────────────────────────────────┘



┌─────────────────────────────────────────────────────────────────────────────────┐

│                            OBSERVABILITY                                         │

│              OpenTelemetry │ Prometheus │ Grafana │ Jaeger                       │

└─────────────────────────────────────────────────────────────────────────────────┘

```



\## Tech Stack



| Layer | Technology | Purpose |

|-------|------------|---------|

| Identity | Keycloak | SSO, OAuth2, JWT |

| Backend (Java) | Spring Boot 3.2 | Order, Warehouse, Auth services |

| Backend (Go) | Go 1.22 + Fiber | Config, Health services |

| Backend (Rust) | Tokio + Axum | GPS Ingestion, API Gateway |

| Backend (Python) | FastAPI | Analytics, ML, Reports |

| Frontend | Next.js 14 | Admin Portal |

| Mobile | React Native + Expo | Driver, Warehouse, Customer apps |

| Database | PostgreSQL 16 | Transactional data |

| Time-Series | TimescaleDB | GPS/Telemetry data |

| Messaging | Apache Kafka | Event streaming |

| Orchestration | Kubernetes | Container orchestration |

| Observability | OpenTelemetry | Distributed tracing |

| Monitoring | Prometheus + Grafana | Metrics \& dashboards |



\## Project Structure



```

.

├── services/

│   ├── auth-service/         # Java - Keycloak integration

│   ├── order-service/        # Java - Order management

│   ├── warehouse-service/    # Java - Inventory management

│   ├── config-service/       # Go - Central configuration

│   ├── health-service/       # Go - Health checks \& metrics

│   ├── gps-ingestion/        # Rust - High-throughput GPS

│   ├── api-gateway/          # Rust - JWT validation \& routing

│   ├── analytics-consumer/   # Python - Kafka to DB

│   ├── ml-pipelines/         # Python - Anomaly detection

│   └── report-generator/     # Python - PDF/Excel reports

├── frontend/

│   ├── admin-portal/         # Next.js dashboard

│   ├── driver-app/           # React Native

│   ├── warehouse-app/        # React Native

│   └── customer-app/         # React Native

├── infrastructure/

│   ├── helm/                 # Helm charts

│   ├── kubernetes/           # K8s manifests

│   ├── monitoring/           # Prometheus, Grafana

│   └── docker/               # Docker configs

├── shared/

│   ├── proto/                # Protobuf definitions

│   └── schemas/              # OpenAPI specs

└── scripts/                  # Utility scripts

```



\## Quick Start



\### Prerequisites



\- Docker \& Docker Compose

\- Kubernetes cluster (minikube/kind for local)

\- Helm 3.x

\- Node.js 20+

\- Go 1.22+

\- Rust 1.75+

\- Python 3.12+

\- Java 21+



\### Local Development



```bash

\# Start infrastructure (Kafka, PostgreSQL, Keycloak, etc.)

make infra-up



\# Start all services

make dev



\# Run tests

make test



\# Build all Docker images

make docker-build



\# Deploy to Kubernetes

make k8s-deploy

```



\### Service Ports



| Service | Port |

|---------|------|

| API Gateway | 8080 |

| Auth Service | 8081 |

| Order Service | 8082 |

| Warehouse Service | 8083 |

| Config Service | 8090 |

| Health Service | 8091 |

| GPS Ingestion | 8100 |

| Analytics Consumer | 8110 |

| ML Pipelines | 8111 |

| Report Generator | 8112 |

| Admin Portal | 3000 |

| Keycloak | 8180 |

| Kafka | 9092 |

| PostgreSQL | 5432 |

| Prometheus | 9090 |

| Grafana | 3001 |

| Jaeger | 16686 |



\## Environment Variables



See `.env.example` for required environment variables.



\## API Documentation



\- OpenAPI specs: `shared/schemas/openapi/`

\- Postman collection: `docs/postman/`



\## License



MIT License



