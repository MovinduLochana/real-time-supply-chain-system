# Logistics Service

The Logistics Service is responsible for managing shipments, tracking locations, and publishing shipment-related events.

## Overview

- **Language**: Rust
- **Port**: 3001
- **Framework**: Axum (HTTP REST API)
- **Database**: PostgreSQL with sqlx
- **Message Queue**: Kafka
- **Observability**: OpenTelemetry with Jaeger

## Features

- Consume OrderCreatedEvent from Kafka and create shipments
- REST APIs for shipment management and location tracking
- Real-time location tracking with GPS coordinates
- Shipment status management
- Publishing ShipmentCreatedEvent, ShipmentStatusChangedEvent, and LocationUpdatedEvent

## Setup

### Prerequisites

- Rust 1.70+
- PostgreSQL 14+
- Kafka 3.0+
- Docker (optional, for containerized deployment)

### Installation

1. Install dependencies:
```bash
cargo build
```

2. Set up environment variables:
```bash
cp .env.example .env
```

3. Run database migrations:
```bash
sqlx migrate run
```

4. Start the service:
```bash
cargo run
```

The service will start on `http://localhost:3001`

## Environment Variables

```
DATABASE_URL=postgresql://user:password@localhost:5432/logistics_db
KAFKA_BROKERS=localhost:9092
KAFKA_CONSUMER_GROUP=logistics-service
RUST_LOG=info,logistics_service=debug
JAEGER_AGENT_HOST=localhost
JAEGER_AGENT_PORT=6831
SERVICE_PORT=3001
```

## API Endpoints

### Create Shipment
**POST** `/shipments`

Request:
```json
{
  "order_id": "order-123",
  "carrier": "FedEx",
  "destination_address": "123 Main St, New York, NY 10001"
}
```

Response:
```json
{
  "shipment_id": "shipment-456",
  "order_id": "order-123",
  "carrier": "FedEx",
  "tracking_number": "1234567890",
  "status": "CREATED",
  "created_at": "2026-05-30T10:00:00Z"
}
```

### Get Shipment
**GET** `/shipments/{shipment_id}`

Response:
```json
{
  "shipment_id": "shipment-456",
  "order_id": "order-123",
  "carrier": "FedEx",
  "tracking_number": "1234567890",
  "status": "IN_TRANSIT",
  "destination_address": "123 Main St, New York, NY 10001",
  "current_location": {
    "latitude": 40.7128,
    "longitude": -74.0060,
    "accuracy_meters": 10.5,
    "address": "New York, NY"
  },
  "created_at": "2026-05-30T10:00:00Z",
  "updated_at": "2026-05-30T12:00:00Z"
}
```

### Record Location Update
**POST** `/shipments/{shipment_id}/locations`

Request:
```json
{
  "latitude": 40.7128,
  "longitude": -74.0060,
  "accuracy_meters": 10.5,
  "address": "New York, NY"
}
```

Response:
```json
{
  "success": true,
  "shipment_id": "shipment-456"
}
```

### Get Location History
**GET** `/shipments/{shipment_id}/locations?limit=50`

Response:
```json
{
  "locations": [
    {
      "latitude": 40.7128,
      "longitude": -74.0060,
      "accuracy_meters": 10.5,
      "address": "New York, NY",
      "recorded_at": "2026-05-30T12:00:00Z"
    }
  ]
}
```

### Update Shipment Status
**PATCH** `/shipments/{shipment_id}/status`

Request:
```json
{
  "status": "IN_TRANSIT",
  "reason": "Package dispatched"
}
```

Response:
```json
{
  "shipment_id": "shipment-456",
  "previous_status": "CREATED",
  "new_status": "IN_TRANSIT",
  "updated_at": "2026-05-30T12:00:00Z"
}
```

## Architecture

### Database Schema

**shipments** table:
- id (UUID, PK)
- order_id (UUID)
- carrier (String)
- tracking_number (String)
- status (Enum)
- destination_address (String)
- current_latitude (Float)
- current_longitude (Float)
- created_at (Timestamp)
- updated_at (Timestamp)

**shipment_locations** table:
- id (UUID, PK)
- shipment_id (UUID, FK)
- latitude (Float)
- longitude (Float)
- accuracy_meters (Float)
- address (String)
- recorded_at (Timestamp)

### Kafka Topics

**Consumed:**
- `order-events`: OrderCreatedEvent

**Published:**
- `shipment-events`: ShipmentCreatedEvent, ShipmentStatusChangedEvent, LocationUpdatedEvent

## Building Docker Image

```bash
docker build -t rtscs/logistics-service:latest .
```

## Deployment

The service is deployed to Kubernetes with the following configuration:
- Health check endpoints: `/health`
- Metrics endpoint: `/metrics`
- Container port: 3001

## Development

### Running Tests
```bash
cargo test
```

### Code Formatting
```bash
cargo fmt
```

### Linting
```bash
cargo clippy
```

## Troubleshooting

### Database Connection Issues
Ensure PostgreSQL is running and `DATABASE_URL` is correctly set:
```bash
sqlx database create
sqlx migrate run
```

### Kafka Connection Issues
Verify Kafka brokers are running and `KAFKA_BROKERS` is set correctly:
```bash
kafka-broker-api-versions.sh --bootstrap-server localhost:9092
```

### Tracing Issues
Ensure Jaeger is running on the configured host/port for telemetry to work.

## License

MIT
