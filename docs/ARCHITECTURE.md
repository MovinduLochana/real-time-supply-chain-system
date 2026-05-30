# RTSCS Architecture & Design

This document outlines the architecture, design decisions, and component relationships in the RTSCS system.

## Overview

RTSCS (Real-Time Supply Chain System) is a modern, polyglot microservices platform designed for high-performance, scalable supply chain management. It uses:

- **gRPC** for fast, typed, synchronous inter-service communication
- **Kafka** for asynchronous, event-driven processing
- **Kubernetes** for orchestration and infrastructure management
- **Next.js** for a responsive, real-time frontend
- **Terraform** for Infrastructure-as-Code (GitOps)
- **Argo CD** for continuous deployment automation

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Users / Partners                         │
│                    (Mobile, Web, Third-party APIs)              │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ REST/JSON + HTTPS
                             │ (JWT Authentication)
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      API Gateway (Bun)                           │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ - REST endpoints (/orders, /inventory, /shipments)     │   │
│  │ - gRPC client pool (to internal services)             │   │
│  │ - WebSocket upgrade handler (real-time updates)       │   │
│  │ - Auth middleware (Keycloak JWT validation)           │   │
│  │ - Request logging, tracing, metrics                   │   │
│  └─────────────────────────────────────────────────────────┘   │
└────┬──────────────────┬──────────────────┬──────────────────────┘
     │                  │                  │
     │ gRPC (mTLS)      │ gRPC (mTLS)      │ gRPC (mTLS)
     │ Port 50051       │ Port 50051       │ Port 50051
     ▼                  ▼                  ▼
┌─────────────────┐ ┌──────────────┐ ┌─────────────────────┐
│ Inventory       │ │ Order        │ │ Logistics           │
│ Service (Java)  │ │ Service (Java)│ │ Service (Rust)      │
├─────────────────┤ ├──────────────┤ ├─────────────────────┤
│ - Stock mgmt    │ │ - Order mgmt │ │ - Shipment tracking │
│ - Reservations  │ │ - Validation │ │ - Location updates  │
│ - Reorder logic │ │ - Workflow   │ │ - GPS processing    │
└────────┬────────┘ └──────┬───────┘ └──────────┬──────────┘
         │                  │                    │
         │ PostgreSQL       │ PostgreSQL         │ PostgreSQL
         │                  │                    │
         └──────────────────┼────────────────────┘
                            │
                      (rtscs_inventory,
                       rtscs_order,
                       rtscs_logistics DBs)
                            │
                      ┌─────▼──────┐
                      │ PostgreSQL  │
                      │ 5432        │
                      └─────┬──────┘
                            │
                            │ Kafka Events Published
                            │ (via Producers)
                            ▼
         ┌──────────────────────────────────────┐
         │        Kafka Event Bus                │
         │ ┌────────────────────────────────┐   │
         │ │ Topics:                        │   │
         │ │ - order-events                 │   │
         │ │ - inventory-events             │   │
         │ │ - shipment-events              │   │
         │ │ - notification-events          │   │
         │ └────────────────────────────────┘   │
         └──────────────────────────────────────┘
                     │        │
                     │        │
        ┌────────────┴┐     ┌─┴──────────────┐
        │ Consumers:  │     │ Consumers:     │
        │ - Notif.    │     │ - Analytics    │
        │ - Redis Sub │     │ - Metrics      │
        ▼             ▼     ▼
   ┌─────────────┐ ┌──────────────┐
   │Notification │ │Analytics     │
   │Service (Go) │ │Service (Py)  │
   ├─────────────┤ ├──────────────┤
   │ - Email     │ │ - Forecasts  │
   │ - SMS       │ │ - Reports    │
   │ - Push      │ │ - Metrics    │
   └────┬────────┘ └───────┬──────┘
        │                  │
        │ Redis Pub/Sub    │ PostgreSQL
        │ (real-time)      │
        └────────┬─────────┘
                 │
                 ▼
         ┌───────────────┐      WebSocket
         │ Redis Pub/Sub │ ────────────────┐
         │ (Cache layer) │                 │
         └───────────────┘                 │
                                           │
                                    ┌──────▼──────────┐
                                    │ Next.js Frontend │
                                    │ Port 3000        │
                                    ├─────────────────┤
                                    │ - Dashboard      │
                                    │ - Order mgmt     │
                                    │ - Real-time map  │
                                    │ - Analytics view │
                                    └──────────────────┘
```

## Data Flow Examples

### Order Creation Flow (Synchronous + Asynchronous)

1. **User** submits order via REST API
2. **API Gateway** validates JWT, translates to gRPC
3. **Order Service** receives `CreateOrder()` call
4. **Order Service** calls `Inventory.ReserveStock()` (gRPC, blocking)
5. If reservation succeeds:
   - Create order in local DB
   - Publish `OrderCreated` event to Kafka (order-events topic)
6. Return order to user
7. **Notification Service** consumes `OrderCreated` event, sends email/SMS
8. **Analytics Service** consumes event for reporting

### Real-Time Shipment Location Update Flow

1. IoT GPS device sends location to `Logistics Service` via gRPC
2. **Logistics Service** processes update, publishes `LocationUpdated` to Kafka
3. **Notification Service** consumes event, publishes to Redis Pub/Sub: `shipment:${id}:location`
4. **API Gateway** subscribed to Redis, broadcasts via WebSocket
5. **Frontend** receives WebSocket message, updates map marker in real-time

## Technology Rationale

### Why gRPC for Internal Communication?

| Aspect | Benefit |
|--------|---------|
| **Performance** | Binary protocol (Protobuf) vs JSON text |
| **Type Safety** | Strongly typed messages, schema evolution with Buf |
| **Polyglot** | First-class support for Java, Go, Rust, Python, TS |
| **Streaming** | Built-in server/client streaming (GPS updates, real-time feeds) |
| **Network** | HTTP/2 multiplexing, connection pooling |
| **Observability** | gRPC interceptors for tracing, metrics, logging |

### Why Kafka for Events?

| Aspect | Benefit |
|--------|---------|
| **Decoupling** | Producers don't wait for consumers |
| **Durability** | Events persisted, replay capability |
| **Scale** | Handles millions of events/sec |
| **Ordering** | Per-partition ordering guarantees |
| **Real-time** | Low-latency pub/sub (< 100ms typical) |

### Why REST at the Edge?

- **Simplicity**: Standard HTTP, easy for clients (web, mobile, third-party APIs)
- **Compatibility**: Works through proxies, firewalls, CDNs
- **Flexibility**: JSON is human-readable for debugging, tooling widely available
- **Gateway translates**: Internal gRPC stays fast and typed, external clients get REST

### Why Next.js + WebSocket?

- **Performance**: SSR + client-side hydration
- **Real-time**: WebSocket for live updates (shipment tracking, stock levels)
- **DX**: React components, TypeScript, Next.js ecosystem
- **Deployment**: Docker container, easy K8s integration

## Service Boundaries

### Inventory Service
**Responsibility**: Manage product catalog, stock levels, and reservations
**gRPC Port**: 50051
**Owns DB**: `rtscs_inventory`
**Kafka Publishes**: `inventory-events` (StockUpdated, StockReserved, LowStockAlert)
**Kafka Consumes**: None (synchronous only)

### Order Service
**Responsibility**: Create, manage, and orchestrate orders
**gRPC Port**: 50051
**Owns DB**: `rtscs_order`
**Kafka Publishes**: `order-events` (OrderCreated, OrderConfirmed, OrderCancelled)
**Kafka Consumes**: `order-events` (for state transitions)
**Calls**: Inventory.ReserveStock()

### Logistics Service
**Responsibility**: Track shipments, handle GPS updates, manage delivery
**gRPC Port**: 50051
**Owns DB**: `rtscs_logistics`
**Kafka Publishes**: `shipment-events` (ShipmentCreated, LocationUpdated, Delivered)
**Kafka Consumes**: `order-events` (to create shipments)
**Specialization**: High-performance Rust for GPS streaming and location processing

### Notification Service
**Responsibility**: Send email, SMS, push notifications, in-app messages
**gRPC Port**: 50051
**Owns DB**: `rtscs_notification` (audit log)
**Kafka Publishes**: `notification-events` (NotificationSent, NotificationFailed)
**Kafka Consumes**: `order-events`, `shipment-events` (trigger-based)
**Redis**: Publishes user-specific updates to Redis Pub/Sub for WebSocket delivery

### Analytics Service
**Responsibility**: Aggregate metrics, generate forecasts, provide reporting
**gRPC/HTTP Port**: 50051 / 8000
**Owns DB**: `rtscs_analytics` (aggregations, forecasts)
**Kafka Publishes**: None (analytics-only output via REST/gRPC)
**Kafka Consumes**: All events (for data warehouse processing)
**Note**: Stateless processing, can auto-scale on Kafka consumer lag

### API Gateway
**Responsibility**: HTTP/REST endpoint, gRPC client aggregation, WebSocket management
**HTTP Port**: 8000
**Owns DB**: None (stateless)
**Calls**: All services via gRPC (Inventory, Order, Logistics, Notification, Analytics)
**Redis**: Subscribes to Redis Pub/Sub for real-time updates, broadcasts via WebSocket

## Communication Patterns

### Synchronous Request/Response (gRPC)
Used for queries and state-changing operations that must succeed immediately:
- Order.CreateOrder() → Inventory.ReserveStock()
- API.GetOrder() → Order.GetOrder()
- Suitable for: Immediate response required, distributed transactions

### Asynchronous Event Publishing (Kafka)
Used for notifications and eventual consistency:
- Order Service → Publishes OrderCreated
- Notification Service → Consumes OrderCreated → sends emails
- Suitable for: Decoupled workflows, multiple consumers, replay capability

### Real-time Updates (WebSocket + Redis Pub/Sub)
Used for live frontend updates:
- Logistics Service → Publishes LocationUpdated → Kafka
- Notification Service → Kafka consumer → Redis Pub/Sub
- API Gateway → WebSocket subscriber → Frontend push
- Suitable for: Real-time dashboards, live tracking, instant notifications

## Deployment Model

- **Local Development**: Kind cluster (3-node) + Terraform IaC
- **Production-Ready**: EKS/GKE/AKS + Terraform
- **GitOps**: Argo CD syncs `/infra/k8s/` changes automatically
- **Infrastructure as Code**: Terraform provisions all services (databases, Kafka, etc.)

## Observability Strategy

- **Tracing**: OpenTelemetry → Jaeger (distributed tracing across gRPC calls)
- **Metrics**: Prometheus scrapes gRPC endpoints, application counters
- **Logging**: Structured JSON logs → Loki (full-text search in Grafana)
- **Dashboards**: Grafana visualizes metrics, traces, logs
- **Alerting**: Prometheus rules → alert on SLO violations

## Security Layers

1. **External**: TLS (HTTPS), JWT validation at API Gateway
2. **Service-to-Service**: mTLS (mutual TLS) between gRPC services via cert-manager
3. **Network**: Kubernetes NetworkPolicy restricts traffic
4. **Data**: PostgreSQL credentials in K8s Secrets
5. **Audit**: Event log in database + Kafka for compliance

## Scaling Strategy

| Component | Scaling Method | Trigger |
|-----------|----------------|---------|
| Order Service | HPA (CPU/Memory) | High request rate |
| Inventory Service | HPA | High concurrent queries |
| Logistics Service | HPA + Kafka consumer lag | Location update volume |
| Notification Service | Kafka consumer groups | Kafka lag |
| Analytics Service | CronJob + stateless processing | Scheduled aggregations |
| PostgreSQL | Vertical scaling (compute) | CPU/memory saturation |
| Redis | Vertical + cluster replication | Memory usage |
| Kafka | Add brokers + rebalance | Topic partition lag |

## Future Enhancements (Phase 5+)

- **Istio Service Mesh**: Advanced traffic management, canary deployments
- **Event Sourcing**: Full event log replay capability
- **CQRS**: Separate read/write models for analytics
- **Multi-region**: Geo-distributed deployments with cross-region replication
- **GraphQL**: Alternative API query language layer
- **Machine Learning**: Real-time demand forecasting with online learning
- **Audit/Compliance**: Full transaction audit trail with immutable log

---

**Document Version**: 1.0  
**Last Updated**: May 30, 2026  
**Author**: Engineering Team
