# Protobuf Guide

This guide explains how to work with Protobuf (.proto files) in RTSCS, from defining messages to generating code.

## Overview

All RTSCS services communicate via gRPC and events through Kafka. Both use Protobuf messages defined in `/proto/v1/` files.

**Why Protobuf?**
- **Schema as contract**: Services agree on message format upfront
- **Backward/forward compatible**: Evolve APIs without breaking clients
- **Type-safe**: Compile-time type checking across languages
- **Efficient**: Binary format (smaller, faster than JSON)
- **Polyglot**: Generate code for Java, Go, Rust, Python, TypeScript automatically

## Directory Structure

```
proto/
├── buf.yaml                    # Buf configuration (lint, breaking change rules)
├── buf.lock                    # Dependency lock file (auto-generated)
├── v1/
│   ├── common/                 # Shared messages
│   │   ├── error.proto         # Error codes and structures
│   │   ├── user.proto          # User context and auth
│   │   └── timestamps.proto    # Timestamp standards
│   │
│   ├── inventory/              # Inventory Service API
│   │   ├── messages.proto      # Item, Stock, Reservation messages
│   │   └── service.proto       # InventoryService gRPC definition
│   │
│   ├── order/                  # Order Service API
│   │   ├── messages.proto      # Order, LineItem messages
│   │   └── service.proto       # OrderService gRPC definition
│   │
│   ├── logistics/              # Logistics Service API
│   │   ├── location.proto      # Shipment, Location, TrackingEvent
│   │   └── service.proto       # LogisticsService gRPC definition
│   │
│   ├── notification/           # Notification Service API
│   │   └── notification.proto  # Notification message and service
│   │
│   ├── analytics/              # Analytics Service API
│   │   └── analytics.proto     # Reports and forecasts
│   │
│   └── events/                 # Kafka event schemas
│       ├── order_events.proto      # Order-related events
│       ├── inventory_events.proto   # Inventory-related events
│       ├── shipment_events.proto    # Shipment-related events
│       └── notification_events.proto # Notification events
│
└── generated/                  # Output directory (auto-generated, gitignored)
    ├── go/
    ├── java/
    ├── python/
    ├── typescript/
    └── rust/
```

## Creating New Messages

### 1. Define Message in .proto File

Open the appropriate `.proto` file and add your message:

```protobuf
syntax = "proto3";

package rtscs.order.v1;

option java_package = "com.rtscs.proto.order.v1";
option go_package = "github.com/rtscs/proto-go/gen/go/rtscs/order/v1";

import "google/protobuf/timestamp.proto";

// Order status enumeration
enum OrderStatus {
  STATUS_UNSPECIFIED = 0;
  PENDING = 1;
  CONFIRMED = 2;
  SHIPPED = 3;
  DELIVERED = 4;
  CANCELLED = 5;
}

// Order message
message Order {
  string order_id = 1;        // Unique order identifier
  string customer_id = 2;     // Customer reference
  double total_amount = 3;    // Order total in USD
  OrderStatus status = 4;     // Current status
  google.protobuf.Timestamp created_at = 5;  // Creation time
}
```

### 2. Proto Syntax Rules

**Field numbering**:
```protobuf
message Item {
  string name = 1;           // Field number (1-536,870,911)
  double price = 2;          // Don't reuse numbers!
  string sku = 3;
}
```

**Reserved fields** (when removing fields):
```protobuf
message Order {
  reserved 5, 6, 7;          // Never reuse these numbers
  reserved "internal_id";     // Never reuse this field name
}
```

**Enumerations**:
```protobuf
enum DeliveryStatus {
  DELIVERY_STATUS_UNSPECIFIED = 0;  // Always include 0 for unspecified
  IN_TRANSIT = 1;
  DELIVERED = 2;
  FAILED = 3;
}
```

**Nested messages**:
```protobuf
message Order {
  message LineItem {
    string sku = 1;
    int32 quantity = 2;
    double price = 3;
  }
  repeated LineItem items = 1;  // repeated = array/list
}
```

## Creating New Services

Services define gRPC endpoints (RPC methods):

```protobuf
syntax = "proto3";

package rtscs.order.v1;

import "rtscs/order/v1/messages.proto";

service OrderService {
  // CreateOrder creates a new order
  rpc CreateOrder (CreateOrderRequest) returns (Order);
  
  // GetOrder retrieves an existing order
  rpc GetOrder (GetOrderRequest) returns (Order);
  
  // StreamOrders returns a stream of orders (server streaming)
  rpc ListOrders (ListOrdersRequest) returns (stream Order);
}

message CreateOrderRequest {
  string customer_id = 1;
  repeated LineItem items = 2;
  
  message LineItem {
    string sku = 1;
    int32 quantity = 2;
  }
}

message GetOrderRequest {
  string order_id = 1;
}

message ListOrdersRequest {
  int32 page = 1;
  int32 page_size = 2;
}
```

**RPC Types**:
- `rpc Method(Request) returns (Response)` — Unary (request/response)
- `rpc Method(Request) returns (stream Response)` — Server streaming
- `rpc Method(stream Request) returns (Response)` — Client streaming
- `rpc Method(stream Request) returns (stream Response)` — Bidirectional streaming

## Validating Changes

### Linting

Buf enforces style rules (Google style by default):

```bash
buf lint proto/
```

Fixes common issues:
- Naming conventions (snake_case for fields)
- Proper imports
- Field numbering

### Breaking Change Detection

Before committing, check you haven't broken existing consumers:

```bash
buf breaking proto/ --against-input='.git#branch=main'
```

**Breaking changes** include:
- Removing a message or field
- Changing field type (int32 → string)
- Changing RPC method signature

**Non-breaking**:
- Adding new fields (old code ignores them)
- Adding new RPC methods
- Adding new messages
- Renaming comments (not field names)

## Generating Code

After modifying `.proto` files:

```bash
bash scripts/gen-proto.sh
```

Or using make:

```bash
make proto-gen
```

This generates:
- **Java**: `services/*/src/main/java/com/rtscs/proto/gen/`
- **Go**: `services/notification-service/pkg/gen/`
- **Rust**: `services/logistics-service/src/gen/`
- **Python**: `services/analytics-service/gen/`
- **TypeScript**: `services/api-gateway/src/gen/`

## Using Generated Code

### Java Example

```java
import com.rtscs.proto.order.v1.*;

// Create a message
Order order = Order.newBuilder()
  .setOrderId("ORD-12345")
  .setCustomerId("CUST-789")
  .setTotalAmount(99.99)
  .setStatus(OrderStatus.PENDING)
  .build();

// Serialize to bytes
byte[] bytes = order.toByteArray();

// Deserialize from bytes
Order restored = Order.parseFrom(bytes);
```

### Go Example

```go
import (
  orderv1 "github.com/rtscs/proto-go/gen/go/rtscs/order/v1"
)

// Create a message
order := &orderv1.Order{
  OrderId:    "ORD-12345",
  CustomerId: "CUST-789",
  TotalAmount: 99.99,
  Status:     orderv1.OrderStatus_PENDING,
}

// Use in gRPC server
func (s *Server) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.Order, error) {
  // ...
  return order, nil
}
```

### Rust Example

```rust
use rtscs_proto::rtscs::order::v1::Order;

// Create a message
let order = Order {
  order_id: "ORD-12345".to_string(),
  customer_id: "CUST-789".to_string(),
  total_amount: 99.99,
  status: order::Status::Pending as i32,
  ..Default::default()
};

// Use in tonic server
#[tonic::async_trait]
impl order_service_server::OrderService for MyOrderService {
  async fn get_order(
    &self,
    request: Request<GetOrderRequest>,
  ) -> Result<Response<Order>, Status> {
    // ...
    Ok(Response::new(order))
  }
}
```

### Python Example

```python
from rtscs.order.v1 import messages_pb2

# Create a message
order = messages_pb2.Order()
order.order_id = "ORD-12345"
order.customer_id = "CUST-789"
order.total_amount = 99.99
order.status = messages_pb2.OrderStatus.PENDING

# Use in gRPC server
class OrderService(service_pb2_grpc.OrderServiceServicer):
  def GetOrder(self, request, context):
    return order
```

### TypeScript Example

```typescript
import { Order, OrderStatus } from '@/gen/rtscs/order/v1/messages_pb';

// Create a message
const order = new Order({
  orderId: "ORD-12345",
  customerId: "CUST-789",
  totalAmount: 99.99,
  status: OrderStatus.PENDING,
});

// Serialize
const bytes = order.toBinary();

// Deserialize
const restored = Order.fromBinary(bytes);
```

## Common Patterns

### Optional Fields (Proto3)

In Proto3, all fields are optional by default:

```protobuf
message Order {
  string order_id = 1;       // Required (but no enforcement in proto3)
  optional string note = 2;  // Explicitly optional
  string customer_id = 3;    // Also optional (language specific)
}
```

For **required fields**, use message-level validation in generated code.

### Default Values

Proto3 auto-initializes:
- Strings: ""
- Numbers: 0
- Booleans: false
- Messages: null/empty

Check for explicit "unset" using `oneof`:

```protobuf
message Order {
  oneof optional_status {
    string status = 1;
    bool status_explicitly_set = 2;
  }
}
```

### Map Types

```protobuf
message Order {
  map<string, int32> item_quantities = 1;  // Map from SKU to quantity
}
```

Generated as:
- Java: `Map<String, Integer>`
- Go: `map[string]int32`
- Python: `dict`

### Timestamps

Always use `google.protobuf.Timestamp`:

```protobuf
import "google/protobuf/timestamp.proto";

message Order {
  google.protobuf.Timestamp created_at = 1;
  google.protobuf.Timestamp updated_at = 2;
}
```

Provides:
- RFC 3339 format
- Timezone-aware (always UTC)
- Language-specific conversions

### Well-Known Types

Use official Protobuf types:

```protobuf
import "google/protobuf/empty.proto";
import "google/protobuf/duration.proto";
import "google/protobuf/any.proto";

service OrderService {
  rpc DeleteOrder(DeleteOrderRequest) returns (google.protobuf.Empty);
}

message Order {
  google.protobuf.Duration processing_time = 1;
}
```

## Workflow

### Adding a New Field (Non-Breaking)

1. Edit `.proto` file, **add new field**:
   ```protobuf
   message Order {
     string order_id = 1;
     string customer_id = 2;
     double total_amount = 3;
     string tracking_url = 4;  // NEW FIELD
   }
   ```

2. Validate:
   ```bash
   buf lint proto/
   buf breaking proto/ --against-input='.git#branch=main'
   ```

3. Generate code:
   ```bash
   bash scripts/gen-proto.sh
   ```

4. Update service code to handle new field

5. Commit and push

### Removing a Field (Breaking Change)

**Don't** directly delete fields! Instead:

1. Mark as deprecated:
   ```protobuf
   message Order {
     reserved 4;                           // Reserve the number
     string order_id = 1;
     string customer_id = 2;
     double total_amount = 3;
     // deprecated string tracking_url = 4;  // NOTE: Removed in v2
   }
   ```

2. Create a new version or major release
3. Communicate deprecation to consumers
4. After sufficient time, remove the field

### Evolving a Service

```protobuf
service OrderService {
  // Existing method (keep for compatibility)
  rpc GetOrder (GetOrderRequest) returns (Order);
  
  // New method (add, don't modify existing)
  rpc GetOrderV2 (GetOrderRequestV2) returns (OrderV2);
  
  // Eventually deprecate v1, but keep it for backward compat
}
```

## CI/CD Integration

GitHub Actions automatically:
1. **Lint** on PR: `buf lint proto/`
2. **Check breaking changes** on PR: `buf breaking`
3. **Generate code** and upload artifact
4. **Validate** proto files match service implementations

See `.github/workflows/ci-proto.yml` for details.

## Testing Proto Changes

### Unit Test (Proto Itself)

Use `buf breaking` to verify backward compatibility:

```bash
buf breaking proto/ --against-input='.git#branch=develop'
```

### Integration Test (Generated Code)

After generation, run language-specific tests:

```bash
# Java
cd services/inventory-service && gradle test

# Go
cd services/notification-service && go test ./...

# Python
cd services/analytics-service && pytest tests/

# Rust
cd services/logistics-service && cargo test
```

## Troubleshooting

### "buf: command not found"

```bash
# Install Buf
brew install buf  # macOS
# Or see scripts/install-tools.sh
```

### Generated Code Not Matching Service

```bash
# Regenerate
bash scripts/gen-proto.sh

# Verify generated files are in .gitignore
cat .gitignore | grep "generated/"
```

### Circular Import Error

```protobuf
// ❌ BAD: circular import
// order.proto imports inventory.proto
// inventory.proto imports order.proto

// ✓ GOOD: create common.proto
// Both import common.proto
```

### Breaking Change Detected But You Didn't Change Anything

```bash
# Check git status
git status

# You may have a git stash or untracked changes
git stash list
```

---

**Document Version**: 1.0  
**Last Updated**: May 30, 2026
