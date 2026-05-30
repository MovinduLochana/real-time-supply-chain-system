# Order Service

Order orchestration service that integrates with Inventory Service via gRPC and publishes events to Kafka.

## Features

- **Order Management**: Create, retrieve, and cancel orders
- **gRPC Client**: Calls Inventory Service to reserve stock
- **gRPC Server**: Exposes OrderService for other services
- **Kafka Producer**: Publishes order lifecycle events
- **Distributed Transactions**: Handles saga pattern for order creation
- **PostgreSQL**: Stores orders and line items
- **OpenTelemetry**: Distributed tracing across services

## Building

```bash
# Build
gradle build

# Docker
docker build -t order-service:latest .

# Run locally
gradle bootRun --args='--spring.profiles.active=dev'
```

## Configuration

### Environment Variables

```bash
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=rtscs_order
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres

KAFKA_BROKERS=localhost:9092

# gRPC client config
INVENTORY_SERVICE_HOST=localhost
INVENTORY_SERVICE_PORT=50051
```

## API

### gRPC Endpoints

- `CreateOrder(customer_id, items, shipping_address) → Order`
- `GetOrder(order_id) → Order`
- `ListOrders(customer_id) → OrderList`
- `CancelOrder(order_id, reason) → Order`

### Example gRPC Call

```bash
grpcurl -plaintext \
  -d '{
    "customerId": "CUST-123",
    "lineItems": [
      {"sku": "SKU-ABC", "quantity": 2, "unitPrice": 29.99}
    ],
    "shippingAddress": "123 Main St, Springfield, IL"
  }' \
  localhost:50051 \
  rtscs.order.v1.OrderService/CreateOrder
```

## Workflow

### Order Creation Flow

1. **Receive CreateOrder request**
   - Validate input
   - Generate order ID

2. **Call Inventory Service (gRPC)**
   - ReserveStock for all items
   - If fails → return error
   - If succeeds → proceed

3. **Create Order in Database**
   - Save OrderEntity
   - Save OrderLineItems
   - Set status to PENDING

4. **Publish OrderCreatedEvent (Kafka)**
   - Topic: order-events
   - Event includes order details
   - For async processing (notifications, analytics)

5. **Return Order to Client**

### Order Cancellation Flow

1. Receive CancelOrder request
2. Update order status to CANCELLED
3. Call Inventory Service to release reservations
4. Publish OrderCancelledEvent
5. Return updated order

## Testing

```bash
# Unit tests
gradle test

# Integration tests (with real services)
gradle integrationTest -i

# Test order creation end-to-end
gradle bootRun &
cd ../../ && make setup-kind && terraform apply...
# Then test with grpcurl
```

## Database

### Schema

```sql
-- Orders
orders
  ├── id (UUID, PK)
  ├── customer_id
  ├── status
  ├── total_amount
  ├── shipping_address
  ├── created_at
  └── updated_at

-- Order Line Items
order_line_items
  ├── id (BIGINT, PK)
  ├── order_id (FK)
  ├── item_sku
  ├── quantity
  └── unit_price
```

## Deployment

```bash
# Local
docker-compose up

# Kubernetes
kubectl apply -f infra/k8s/apps/order/

# Verify
kubectl get pods -n rtscs-services
kubectl logs -f deployment/order-service -n rtscs-services
```

## Monitoring

### Logs

```bash
kubectl logs -f deployment/order-service -n rtscs-services
```

### Metrics

```bash
curl http://localhost:8080/actuator/prometheus | grep order
```

### Traces

Open Jaeger: http://localhost:16686
- Search for traces from "order-service"
- View gRPC calls to "inventory-service"
- Check latency and dependencies

## Troubleshooting

### gRPC Connection to Inventory Service

```bash
grpcurl -plaintext localhost:50051 list
```

### Kafka Events Not Publishing

```bash
# Check topic
kafka-topics.sh --bootstrap-server localhost:9092 --list

# Monitor events
kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic order-events \
  --from-beginning
```

### Database Transactions

```bash
psql -h localhost -U postgres -d rtscs_order \
  -c "SELECT id, customer_id, status FROM orders LIMIT 10;"
```

## Performance

### Optimization

- **Connection pooling** to Inventory Service
- **Batch inserts** via Hibernate
- **Async Kafka** publishing
- **Index on customer_id and status**

### Latency SLO

- CreateOrder: < 500ms (including gRPC call to Inventory)
- GetOrder: < 50ms
- CancelOrder: < 200ms

## Related Services

- **Inventory Service**: Provides stock reservation via gRPC
- **Notification Service**: Consumes order-events, sends notifications
- **Analytics Service**: Consumes order-events for reporting
- **Logistics Service**: Receives OrderCreatedEvent, creates shipments

## License

See LICENSE file in repository root.
