# Notification Service

A high-performance notification service for RTSCS (Phase 2) that consumes events from Kafka and delivers notifications via email (SendGrid), SMS (Twilio), and push notifications (Firebase Cloud Messaging).

## Overview

The Notification Service is responsible for:

- Consuming events from Kafka topics (`order-events`, `inventory-events`)
- Sending notifications via multiple channels (email, SMS, push)
- Managing user notification preferences
- Tracking notification status and delivery
- Publishing notification events back to Kafka
- Handling retries with exponential backoff
- Providing REST APIs for direct notification sending

## Architecture

### Components

- **Kafka Consumer**: Listens to order and inventory events
- **Kafka Producer**: Publishes notification status events
- **Email Service**: Sends emails via SendGrid API
- **SMS Service**: Sends SMS via Twilio API
- **Push Service**: Sends push notifications via Firebase Cloud Messaging
- **Notification Service**: Core business logic with worker pool pattern
- **Redis**: Persistent storage and task queue
- **HTTP API**: REST endpoints for notification management

### Data Flow

```
Kafka Events → Consumer → Event Handlers → Notification Service
                                              ↓
                                        Worker Pool (async)
                                              ↓
                                    Email/SMS/Push Services
                                              ↓
                                    Redis (store status)
                                              ↓
                                    Kafka Producer (publish events)
```

## Prerequisites

- Go 1.21+
- Redis (6.0+)
- Kafka (2.8+)
- SendGrid API key
- Twilio account credentials
- Firebase project (optional, for push notifications)

## Installation

### 1. Clone and Setup

```bash
cd services/notification-service
go mod download
```

### 2. Environment Configuration

Create a `.env` file in the service root:

```env
# Server
SERVER_PORT=3002
ENVIRONMENT=development
LOG_LEVEL=info

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_CONSUMER_GROUP=notification-service
KAFKA_TOPICS=order-events,inventory-events
KAFKA_READ_TIMEOUT_SECONDS=30

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_MAX_RETRIES=3
REDIS_TTL_DAYS=30

# SendGrid
SENDGRID_API_KEY=your-sendgrid-api-key

# Twilio
TWILIO_ACCOUNT_SID=your-account-sid
TWILIO_AUTH_TOKEN=your-auth-token
TWILIO_FROM_NUMBER=+1234567890

# Firebase Cloud Messaging
FCM_PROJECT_ID=your-project-id
FCM_KEY_PATH=/path/to/serviceAccountKey.json

# OpenTelemetry
OTEL_ENABLED=false
OTEL_ENDPOINT=localhost:4318

# Service Configuration
WORKER_POOL_SIZE=10
MAX_RETRIES=3
RETRY_WAIT_SECONDS=2
HTTP_TIMEOUT_SECONDS=30
```

## Running Locally

### Development

```bash
# Install dependencies
go mod download

# Run the service
go run main.go
```

The service will start on `http://localhost:3002`

### With Docker

```bash
# Build Docker image
docker build -t notification-service:latest .

# Run container
docker run -p 3002:3002 \
  -e KAFKA_BROKERS=kafka:9092 \
  -e REDIS_ADDR=redis:6379 \
  -e SENDGRID_API_KEY=your-key \
  -e TWILIO_ACCOUNT_SID=your-sid \
  -e TWILIO_AUTH_TOKEN=your-token \
  notification-service:latest
```

### Docker Compose

```yaml
version: '3.8'

services:
  notification-service:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "3002:3002"
    environment:
      KAFKA_BROKERS: kafka:9092
      REDIS_ADDR: redis:6379
      SENDGRID_API_KEY: ${SENDGRID_API_KEY}
      TWILIO_ACCOUNT_SID: ${TWILIO_ACCOUNT_SID}
      TWILIO_AUTH_TOKEN: ${TWILIO_AUTH_TOKEN}
    depends_on:
      - kafka
      - redis
```

## API Documentation

### Health Check

**Endpoint**: `GET /health`

**Response** (200 OK):
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "services": {
    "redis": {
      "status": "healthy"
    }
  }
}
```

### Send Email Notification

**Endpoint**: `POST /notifications/email`

**Request**:
```json
{
  "user_id": "user-123",
  "recipient_id": "user@example.com",
  "subject": "Order Confirmation",
  "body": "Your order has been confirmed",
  "metadata": {
    "order_id": "order-456"
  }
}
```

**Response** (202 Accepted):
```json
{
  "notification_id": "notif-uuid",
  "status": "PENDING"
}
```

### Send SMS Notification

**Endpoint**: `POST /notifications/sms`

**Request**:
```json
{
  "user_id": "user-123",
  "recipient_id": "+1234567890",
  "body": "Your order has been shipped",
  "metadata": {
    "order_id": "order-456"
  }
}
```

**Response** (202 Accepted):
```json
{
  "notification_id": "notif-uuid",
  "status": "PENDING"
}
```

### Send Push Notification

**Endpoint**: `POST /notifications/push`

**Request**:
```json
{
  "user_id": "user-123",
  "recipient_id": "device-id-123",
  "subject": "Order Update",
  "body": "Your order is on the way",
  "metadata": {
    "order_id": "order-456"
  }
}
```

**Response** (202 Accepted):
```json
{
  "notification_id": "notif-uuid",
  "status": "PENDING"
}
```

### Get Notification Status

**Endpoint**: `GET /notifications/{id}`

**Response** (200 OK):
```json
{
  "id": "notif-uuid",
  "user_id": "user-123",
  "channel": "email",
  "recipient_id": "user@example.com",
  "subject": "Order Confirmation",
  "body": "Your order has been confirmed",
  "status": "SENT",
  "metadata": {
    "order_id": "order-456"
  },
  "retry_count": 0,
  "max_retries": 3,
  "created_at": "2024-01-01T00:00:00Z",
  "sent_at": "2024-01-01T00:00:05Z",
  "delivered_at": null,
  "failed_at": null,
  "error_message": "",
  "correlation_id": "corr-uuid"
}
```

### Set User Preferences

**Endpoint**: `POST /notifications/preferences/{user_id}`

**Request**:
```json
{
  "email_opt_in": true,
  "sms_opt_in": true,
  "push_opt_in": false,
  "do_not_disturb": false,
  "do_not_disturb_end": null,
  "channels": {
    "email": "user@example.com",
    "sms": "+1234567890",
    "push": "device-id-123"
  }
}
```

**Response** (200 OK):
```json
{
  "user_id": "user-123",
  "message": "preferences updated"
}
```

## Event Models

### OrderCreatedEvent (Kafka)

Topic: `order-events`

```json
{
  "order_id": "order-123",
  "user_id": "user-456",
  "order_amount": 99.99,
  "created_at": "2024-01-01T00:00:00Z",
  "correlation_id": "corr-uuid"
}
```

**Action**: Creates email confirmation notification

### OrderCancelledEvent (Kafka)

Topic: `order-events`

```json
{
  "order_id": "order-123",
  "user_id": "user-456",
  "cancel_reason": "Customer requested cancellation",
  "cancelled_at": "2024-01-01T00:00:00Z",
  "correlation_id": "corr-uuid"
}
```

**Action**: Creates email cancellation notification

### LowStockAlertEvent (Kafka)

Topic: `inventory-events`

```json
{
  "product_id": "product-123",
  "product_name": "Widget",
  "current_stock": 5,
  "alerted_at": "2024-01-01T00:00:00Z",
  "correlation_id": "corr-uuid"
}
```

**Action**: Creates email low stock alert

### NotificationSentEvent (Published)

Topic: `notification-events`

```json
{
  "notification_id": "notif-uuid",
  "user_id": "user-123",
  "channel": "email",
  "sent_at": "2024-01-01T00:00:05Z",
  "correlation_id": "corr-uuid"
}
```

### NotificationFailedEvent (Published)

Topic: `notification-events`

```json
{
  "notification_id": "notif-uuid",
  "user_id": "user-123",
  "channel": "email",
  "error_message": "Invalid email address",
  "failed_at": "2024-01-01T00:00:10Z",
  "correlation_id": "corr-uuid"
}
```

## Redis Data Schema

See `migrations/redis-schema.txt` for complete Redis data structure documentation.

### Key Patterns

- `notification:{id}` - Individual notification data (TTL: 30 days)
- `user-preferences:{user_id}` - User notification preferences
- `notification-queue:{channel}` - Task queue for pending notifications

## Configuration

### Worker Pool

- **WORKER_POOL_SIZE**: Number of concurrent notification workers (default: 10)
- Adjustable based on expected throughput and infrastructure

### Retry Strategy

- **MAX_RETRIES**: Maximum retry attempts per notification (default: 3)
- **RETRY_WAIT_SECONDS**: Initial wait time for exponential backoff (default: 2)
- Backoff formula: `backoff_seconds = 2^retry_count * RETRY_WAIT_SECONDS`
- Retries: 2s → 4s → 8s → fail

### TTL & Cleanup

- Notifications expire after **REDIS_TTL_DAYS** (default: 30 days)
- Automatic cleanup via Redis TTL
- Manual cleanup available via admin API

## Monitoring

### Logs

Service logs follow structured logging format with correlation IDs for tracing:

```
INFO  HTTP request              method=POST path=/notifications/email status=202 duration_ms=45 request_id=req-123 correlation_id=corr-456
INFO  Notification sent         notification_id=notif-789 user_id=user-123 channel=email correlation_id=corr-456
```

### Metrics

Key metrics to monitor:

- Notification queue lengths per channel
- Delivery success/failure rates
- Retry counts and backoff times
- API response times
- Worker pool utilization
- Redis connection health

### OpenTelemetry

Enable distributed tracing:

```env
OTEL_ENABLED=true
OTEL_ENDPOINT=localhost:4318
```

Traces include:
- HTTP requests
- Kafka message processing
- Database operations
- Notification delivery attempts

## Error Handling

### Notification Failures

Failed notifications are:
1. Retried with exponential backoff (up to MAX_RETRIES)
2. Logged with error details
3. Marked as FAILED after max retries
4. Published as NotificationFailedEvent
5. Stored in Redis for 30 days

### User Preference Handling

- Invalid/missing recipients → notification marked FAILED
- User opted out → notification marked UNSUBSCRIBED
- Do not disturb active → notification queued for later
- Missing preferences → defaults used (all channels opted in)

## Production Deployment

### Recommended Configuration

```env
# High-throughput settings
WORKER_POOL_SIZE=20
MAX_RETRIES=5
RETRY_WAIT_SECONDS=3

# Kubernetes liveness probe
GET /health every 30s with 3s timeout

# Resource limits
Memory: 256Mi - 512Mi
CPU: 250m - 1000m

# Redis cluster recommended for HA
REDIS_ADDR=redis-cluster:6379

# Kafka broker replication
KAFKA_BROKERS=broker1:9092,broker2:9092,broker3:9092
```

### Security

- SendGrid API key → environment variable (never in code)
- Twilio credentials → environment variable
- Redis password protection if exposed to network
- HTTPS for API endpoints (use reverse proxy)
- Correlation ID for audit logging

## Development

### Project Structure

```
.
├── main.go                    # Entry point
├── config/
│   └── config.go             # Configuration management
├── models/
│   └── notification.go       # Data models
├── services/
│   ├── notification.go       # Core service logic
│   ├── email.go              # SendGrid integration
│   ├── sms.go                # Twilio integration
│   └── push.go               # Firebase integration
├── kafka/
│   ├── consumer.go           # Kafka message consumer
│   └── producer.go           # Kafka message producer
├── db/
│   └── redis.go              # Redis client
├── api/
│   ├── router.go             # HTTP router setup
│   └── handlers/
│       ├── notification.go   # Notification endpoints
│       └── health.go         # Health check endpoint
├── middleware/
│   └── middleware.go         # HTTP middleware
├── errors/
│   └── errors.go             # Error types
└── migrations/
    └── redis-schema.txt      # Schema documentation
```

### Code Standards

- Error handling: explicit with logging
- Logging: structured with zap, include correlation IDs
- Concurrency: goroutines with context cancellation
- Timeouts: on all external service calls
- Testing: unit tests for business logic, integration tests in CI/CD

## Troubleshooting

### Service won't start

1. Check Redis connection: `redis-cli ping`
2. Check Kafka brokers: `kafka-broker-api-versions.sh --bootstrap-server localhost:9092`
3. Check API keys in environment variables
4. Review logs for specific errors

### Notifications not being sent

1. Check user preferences: `GET /notifications/preferences/{user_id}`
2. Check notification status: `GET /notifications/{id}`
3. Verify SendGrid/Twilio credentials
4. Check Redis queue length: `LLEN notification-queue:email`
5. Review service logs for error messages

### High failure rate

1. Verify recipient addresses/phone numbers
2. Check rate limits on SendGrid/Twilio
3. Monitor Redis memory usage
4. Check network connectivity
5. Review correlation ID logs for patterns

## Support

For issues and questions:
- Check logs with correlation ID
- Review Redis data with `redis-cli`
- Check Kafka messages with Kafka tools
- Monitor metrics dashboard

## License

RTSCS Services - Proprietary
