# Inventory Service

High-performance inventory management service built with Java, Spring Boot, and gRPC.

## Features

- **Stock Management**: Track inventory levels across multiple warehouses
- **Reservations**: Reserve stock for orders with automatic expiration
- **gRPC API**: Fast, typed service-to-service communication
- **Kafka Events**: Publish stock changes for async processing
- **PostgreSQL**: Transactional persistence with optimistic locking
- **Distributed Tracing**: OpenTelemetry integration with Jaeger
- **Health Checks**: Kubernetes-ready health endpoints

## Building

### Prerequisites
- Java 21+
- Gradle 8.5+
- PostgreSQL 15+

### Build Commands

```bash
# Build service
gradle build

# Build Docker image
docker build -t inventory-service:latest .

# Run locally (requires PostgreSQL)
gradle bootRun --args='--spring.profiles.active=dev'
```

## Configuration

### Environment Variables

```bash
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=rtscs_inventory
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres

KAFKA_BROKERS=localhost:9092
JAEGER_AGENT_HOST=localhost
JAEGER_AGENT_PORT=6831

LOG_LEVEL=INFO
```

### application.yml

See `src/main/resources/application.yml` for detailed configuration options.

## API

### gRPC Endpoints

- `GetStock(sku, warehouse_location) → StockLevel`
- `UpdateStock(sku, delta, reason) → UpdateStockResponse`
- `ReserveStock(order_id, items) → Reservation`
- `ReleaseReservation(reservation_id) → ReleaseReservationResponse`

### REST Endpoints (Health/Metrics)

- `GET /actuator/health` — Service health
- `GET /actuator/metrics` — Prometheus metrics
- `GET /actuator/metrics/rtscs.*` — Custom metrics

## Testing

```bash
# Unit tests
gradle test

# Integration tests (requires Docker)
gradle integrationTest

# Test coverage
gradle test jacocoTestReport
```

## Database

### Migrations

Managed by Flyway. See `src/main/resources/db/migration/` for schema.

### Running Migrations

```bash
gradle flywayMigrate
```

## Deployment

### Docker Compose (Local)

```bash
docker-compose up
```

### Kubernetes

```bash
kubectl apply -f infra/k8s/apps/inventory/
```

## Monitoring

### Logs

```bash
# Live logs
kubectl logs -f deployment/inventory-service -n rtscs-services

# Structured logs
kubectl logs deployment/inventory-service -n rtscs-services | jq .
```

### Metrics

```bash
# View metrics
curl http://localhost:8080/actuator/metrics | jq .
curl http://localhost:8080/actuator/prometheus
```

### Tracing

Access Jaeger UI: http://localhost:16686
Search for traces from service "inventory-service"

## Troubleshooting

### gRPC Connection Issues

```bash
# Test gRPC endpoint
grpcurl -plaintext localhost:50051 list

# Get method details
grpcurl -plaintext localhost:50051 describe rtscs.inventory.v1.InventoryService
```

### Database Connection

```bash
# Test PostgreSQL
psql -h localhost -U postgres -d rtscs_inventory -c "SELECT COUNT(*) FROM items;"
```

### Kafka Issues

```bash
# Check topic
kafka-topics.sh --bootstrap-server localhost:9092 --list

# Monitor events
kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic inventory-events
```

## Development

### Code Structure

```
src/
├── main/
│   ├── java/com/rtscs/inventory/
│   │   ├── InventoryServiceApplication.java    # Main app
│   │   ├── adapter/grpc/                       # gRPC endpoint
│   │   ├── application/service/                # Business logic
│   │   ├── domain/                             # Entities + repositories
│   │   └── infrastructure/                     # Config, utilities
│   └── resources/
│       ├── application.yml                     # Configuration
│       └── db/migration/                       # SQL migrations
└── test/
    └── java/com/rtscs/inventory/              # Tests
```

### Adding New Features

1. Update `.proto` files in `proto/v1/inventory/`
2. Run `gradle clean build` to regenerate
3. Implement logic in `application/service/`
4. Add gRPC handler in `adapter/grpc/`
5. Write tests in `src/test/`
6. Commit and push

## Performance

### Optimizations

- **Pessimistic locking** on stock updates (prevents race conditions)
- **Connection pooling** (HikariCP, 20 max connections)
- **Caching** (Spring Cache with Redis backend)
- **Batch processing** (Hibernate batch_size=20)
- **Async Kafka** (non-blocking event publishing)

### Benchmarks

Typical latency:
- GetStock: < 10ms (cached)
- UpdateStock: < 50ms (with lock)
- ReserveStock: < 100ms (multiple items)

## License

See LICENSE file in repository root.
