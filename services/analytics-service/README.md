# Analytics Service (Phase 2)

Real-time analytics and forecasting service for RTSCS. Consumes events from Kafka, performs aggregations, and generates demand forecasts.

## Overview

The Analytics Service is a FastAPI-based microservice that:
- Consumes order and inventory events from Kafka
- Performs real-time aggregations and calculations
- Stores metrics in DuckDB for fast analytics
- Generates demand forecasts using time-series models
- Provides REST APIs for metrics, forecasting, and reporting

## Technology Stack

- **Framework**: FastAPI 0.104.1
- **Database**: DuckDB 0.9.2 (in-memory OLAP + parquet persistence)
- **Message Queue**: Apache Kafka (via aio-pika)
- **Async**: asyncio with aio-pika
- **Data Processing**: Pandas, NumPy, Scikit-learn
- **Forecasting**: Prophet 1.1.5 / ARIMA (statsmodels)
- **Observability**: OpenTelemetry with Jaeger
- **Language**: Python 3.11+

## Project Structure

```
services/analytics-service/
├── main.py                    # Application entry point
├── config.py                  # Configuration management
├── requirements.txt           # Python dependencies
├── Dockerfile                 # Container image
├── README.md                  # This file
├── app/
│   ├── __init__.py
│   ├── api/
│   │   ├── __init__.py
│   │   ├── routes.py          # FastAPI routes
│   │   └── handlers.py        # Request handlers
│   ├── kafka/
│   │   ├── __init__.py
│   │   ├── consumer.py        # Kafka event consumer
│   │   └── producer.py        # Kafka event producer
│   ├── analytics/
│   │   ├── __init__.py
│   │   ├── metrics.py         # Metrics calculations
│   │   ├── forecasting.py     # Demand forecasting
│   │   └── aggregations.py    # Data aggregations
│   ├── database/
│   │   ├── __init__.py
│   │   ├── duckdb_client.py   # DuckDB client
│   │   └── queries.py         # SQL queries and schemas
│   ├── models/
│   │   ├── __init__.py
│   │   └── schemas.py         # Pydantic models
│   └── utils/
│       ├── __init__.py
│       ├── logging.py         # Structured logging
│       └── tracing.py         # OpenTelemetry setup
├── migrations/
│   └── init_schema.sql        # Database schema
└── data/
    └── .gitkeep
```

## Installation

### Local Development

```bash
# Clone the repository
cd services/analytics-service

# Create virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Set environment variables
export KAFKA_BROKERS=localhost:9092
export DB_PATH=./data/analytics.duckdb
export OTEL_ENABLED=false

# Run the service
python main.py
```

### Docker

```bash
# Build image
docker build -t analytics-service:latest .

# Run container
docker run -p 3003:3003 \
  -e KAFKA_BROKERS=kafka:9092 \
  -e OTEL_ENABLED=true \
  -e JAEGER_HOST=jaeger \
  -v analytics-data:/app/data \
  analytics-service:latest
```

## Configuration

Configuration is managed through environment variables. See `config.py` for all available options.

### Key Settings

```python
# Service
SERVICE_NAME=analytics-service
SERVICE_PORT=3003
SERVICE_HOST=0.0.0.0
DEBUG=false

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_ORDER_EVENTS_TOPIC=order-events
KAFKA_INVENTORY_EVENTS_TOPIC=inventory-events
KAFKA_METRICS_TOPIC=metrics-events
KAFKA_FORECAST_TOPIC=forecast-events
KAFKA_CONSUMER_GROUP=analytics-service

# Database
DB_PATH=./data/analytics.duckdb
DB_PARQUET_PATH=./data/parquet

# OpenTelemetry
OTEL_ENABLED=false
JAEGER_HOST=localhost
JAEGER_PORT=6831

# Analytics
FORECAST_DAYS_DEFAULT=30
FORECAST_DAYS_MAX=365
METRICS_CACHE_TTL_SECONDS=300
```

Create a `.env` file in the root directory to set these values.

## API Documentation

The service provides comprehensive REST APIs for metrics, forecasting, and reporting.

### Metrics Endpoints

#### Get Order Metrics
```
GET /metrics/orders
```

Returns order statistics for the last 30 days.

**Response:**
```json
{
  "total_orders": 1250,
  "total_revenue": 125000.50,
  "average_order_value": 100.00,
  "order_rate_per_day": 41.67,
  "period_start": "2024-04-30T00:00:00",
  "period_end": "2024-05-30T23:59:59",
  "timestamp": "2024-05-30T15:30:00"
}
```

#### Get Inventory Metrics
```
GET /metrics/inventory
```

Returns current inventory statistics.

**Response:**
```json
{
  "total_skus": 150,
  "total_stock_quantity": 5000,
  "low_stock_count": 12,
  "average_turnover_rate": 33.33,
  "by_sku": {
    "SKU-001": {
      "current_stock": 100,
      "net_change": 15,
      "update_count": 25,
      "last_updated": "2024-05-30T14:22:00"
    }
  },
  "timestamp": "2024-05-30T15:30:00"
}
```

#### Get Customer Metrics
```
GET /metrics/customers
```

Returns customer analytics.

**Response:**
```json
{
  "total_customers": 500,
  "new_customers_this_period": 12,
  "repeat_customer_count": 250,
  "repeat_rate": 0.50,
  "average_clv": 250.00,
  "timestamp": "2024-05-30T15:30:00"
}
```

### Forecasting Endpoint

#### Get Demand Forecast
```
GET /forecast/demand?sku=SKU-001&days=30
```

Generates demand forecast for a SKU using Prophet model.

**Parameters:**
- `sku` (required): Product SKU identifier
- `days` (optional): Number of days to forecast (1-365, default: 30)

**Response:**
```json
{
  "sku": "SKU-001",
  "forecast_days": 30,
  "model_type": "prophet",
  "confidence_level": 0.92,
  "forecast": [
    {
      "date": "2024-05-31",
      "yhat": 45.5,
      "yhat_lower": 40.2,
      "yhat_upper": 50.8,
      "trend": 45.0
    }
  ],
  "generated_at": "2024-05-30T15:30:00"
}
```

### Reports Endpoint

#### Get Daily Report
```
GET /reports/daily?date=2024-05-30
```

Generates comprehensive daily analytics report.

**Parameters:**
- `date` (optional): Report date in YYYY-MM-DD format (defaults to today)

**Response:**
```json
{
  "report_date": "2024-05-30",
  "order_metrics": { ... },
  "inventory_metrics": { ... },
  "customer_metrics": { ... },
  "generated_at": "2024-05-30T15:30:00"
}
```

### Analytics Endpoints

#### Get Daily Trends
```
GET /analytics/daily-trends?days=30
```

Returns daily order trends over time period.

#### Get Low Stock Items
```
GET /analytics/low-stock
```

Returns items currently below stock threshold.

#### Get Top Revenue SKUs
```
GET /analytics/top-skus?limit=10&days=30
```

Returns top revenue-generating SKUs.

**Parameters:**
- `limit` (optional): Number of SKUs to return (1-100, default: 10)
- `days` (optional): Analysis period in days (default: 30)

#### Get Customer Lifetime Value
```
GET /analytics/customer-value?limit=10
```

Returns customers ranked by lifetime value.

### Health Check

```
GET /health
```

Returns service health status and dependency availability.

**Response:**
```json
{
  "status": "healthy",
  "service": "analytics-service",
  "version": "2.0.0",
  "database_connected": true,
  "kafka_connected": true,
  "timestamp": "2024-05-30T15:30:00"
}
```

## Event Processing

### Consumed Events

The service consumes events from Kafka topics and processes them:

#### Order Events (topic: `order-events`)

**OrderCreatedEvent**
```json
{
  "event_type": "OrderCreatedEvent",
  "event_id": "evt-123",
  "order_id": "ORD-001",
  "customer_id": "CUST-001",
  "total_amount": 150.50,
  "item_count": 3,
  "sku_list": ["SKU-001", "SKU-002"],
  "timestamp": "2024-05-30T15:30:00"
}
```

**OrderCancelledEvent**
```json
{
  "event_type": "OrderCancelledEvent",
  "event_id": "evt-124",
  "order_id": "ORD-001",
  "timestamp": "2024-05-30T15:35:00",
  "reason": "Customer requested"
}
```

#### Inventory Events (topic: `inventory-events`)

**StockUpdatedEvent**
```json
{
  "event_type": "StockUpdatedEvent",
  "event_id": "evt-125",
  "sku": "SKU-001",
  "previous_quantity": 100,
  "new_quantity": 95,
  "warehouse_id": "WH-001",
  "timestamp": "2024-05-30T15:30:00"
}
```

**LowStockAlertEvent**
```json
{
  "event_type": "LowStockAlertEvent",
  "event_id": "evt-126",
  "sku": "SKU-001",
  "current_quantity": 5,
  "threshold": 10,
  "timestamp": "2024-05-30T15:30:00"
}
```

### Published Events

The service publishes analytics events:

**MetricsAggregatedEvent** (topic: `metrics-events`)
```json
{
  "event_id": "evt-200",
  "event_type": "MetricsAggregatedEvent",
  "period": "hourly",
  "timestamp": "2024-05-30T16:00:00",
  "metrics": {
    "orders": {
      "count": 45,
      "revenue": 4500.00
    },
    "inventory": {
      "updates": 120,
      "skus_updated": 50
    }
  }
}
```

**ForecastGeneratedEvent** (topic: `forecast-events`)
```json
{
  "event_id": "evt-201",
  "event_type": "ForecastGeneratedEvent",
  "sku": "SKU-001",
  "days": 30,
  "model_type": "prophet",
  "timestamp": "2024-05-30T16:00:00",
  "forecast_data": [
    {
      "date": "2024-05-31",
      "yhat": 45.5,
      "yhat_lower": 40.2,
      "yhat_upper": 50.8
    }
  ]
}
```

## Database Schema

The service uses DuckDB with the following schema:

### orders_fact
Fact table for order events.
- `order_id` (VARCHAR): Primary key
- `customer_id` (VARCHAR): Customer identifier
- `total_amount` (DOUBLE): Order total
- `item_count` (INTEGER): Number of items
- `status` (VARCHAR): created, cancelled
- `created_at` (TIMESTAMP): Order creation time
- `cancelled_at` (TIMESTAMP): Cancellation time
- `sku_list` (VARCHAR[]): Product SKUs
- `event_id` (VARCHAR): Event identifier
- `ingested_at` (TIMESTAMP): Ingestion timestamp

### inventory_fact
Fact table for inventory changes.
- `sku` (VARCHAR): Product SKU
- `warehouse_id` (VARCHAR): Warehouse identifier
- `previous_quantity` (INTEGER): Previous stock level
- `new_quantity` (INTEGER): New stock level
- `updated_at` (TIMESTAMP): Update time
- `alert_threshold` (INTEGER): Low stock threshold
- `is_low_stock` (BOOLEAN): Low stock flag
- `event_id` (VARCHAR): Event identifier
- `ingested_at` (TIMESTAMP): Ingestion timestamp

### customers_dim
Dimension table for customer attributes.
- `customer_id` (VARCHAR): Primary key
- `first_seen` (TIMESTAMP): First activity time
- `last_seen` (TIMESTAMP): Last activity time
- `total_orders` (INTEGER): Order count
- `total_spent` (DOUBLE): Lifetime spend
- `is_repeat_customer` (BOOLEAN): Repeat customer flag
- `updated_at` (TIMESTAMP): Last update time

### forecasts
Forecast results storage.
- `forecast_id` (VARCHAR): Primary key
- `sku` (VARCHAR): Product SKU
- `forecast_days` (INTEGER): Forecast horizon
- `model_type` (VARCHAR): prophet, arima
- `forecast_data` (VARCHAR): JSON forecast points
- `confidence_level` (DOUBLE): Model confidence
- `generated_at` (TIMESTAMP): Generation time
- `valid_until` (TIMESTAMP): Forecast validity end
- `event_id` (VARCHAR): Event identifier

## Analytics Features

### Real-time Metrics

The service calculates real-time metrics:
- Order count, revenue, average order value
- Inventory levels by SKU, turnover rates
- Customer acquisition, retention, lifetime value
- Daily/weekly/monthly trends

### Demand Forecasting

Prophet-based time-series forecasting:
- 7, 30, 90-day forecasts per SKU
- Automatic seasonality detection
- Confidence intervals (95%)
- ARIMA fallback if Prophet unavailable

### Data Aggregation

Automatic data aggregation:
- Hourly metrics for real-time monitoring
- Daily metrics for reporting
- Weekly/monthly summaries
- Low stock alerts

### Data Persistence

- Parquet export for batch analysis
- Data retention policies (90 days default)
- Automatic cleanup of old forecasts

## Logging

The service uses structured JSON logging with fields:
- `service`: Service name
- `level`: Log level
- `logger`: Logger name
- `message`: Log message
- `timestamp`: Log timestamp
- Exception details included for errors

**Log Level**: Set via `LOG_LEVEL` environment variable (default: INFO)

## Monitoring & Observability

### OpenTelemetry Integration

When enabled (`OTEL_ENABLED=true`):
- Distributed tracing with Jaeger
- Automatic span creation for requests
- Custom metrics for service performance
- Resource attributes (service name, version)

### Health Checks

Docker health check endpoint:
```bash
curl http://localhost:3003/health
```

Returns 200 when service is healthy, includes database and Kafka connectivity status.

## Development

### Running Tests

Tests are not included in this implementation but can be added:

```bash
pytest tests/ -v
```

### Code Style

Code follows PEP 8 with:
- Type hints throughout
- Structured logging
- Async/await patterns
- Error handling

### Adding New Metrics

1. Add query to `app/database/queries.py`
2. Create calculation method in `app/analytics/metrics.py`
3. Add endpoint in `app/api/routes.py`
4. Create handler in `app/api/handlers.py`

### Adding New Event Types

1. Define schema in `app/models/schemas.py`
2. Create handler in `app/api/handlers.py`
3. Register handler in `main.py` startup
4. Update Kafka consumer in `app/kafka/consumer.py`

## Performance Considerations

- **DuckDB**: In-memory OLAP for fast analytics queries
- **Parquet**: Columnar storage for efficient batch access
- **Indexing**: Automatic on key columns
- **Metrics Caching**: TTL-based cache to reduce calculations
- **Async Processing**: Non-blocking event processing
- **Batch Operations**: Batched Kafka consumption

## Security

- CORS enabled for all origins (configure as needed)
- No authentication currently implemented (add as needed)
- Sensitive config in environment variables
- Structured logging without sensitive data exposure

## Troubleshooting

### Kafka Connection Issues
- Check `KAFKA_BROKERS` setting
- Verify Kafka broker is running
- Check network connectivity

### Database Issues
- Verify `DB_PATH` directory is writable
- Check available disk space
- Review DuckDB error logs

### Forecasting Failures
- Ensure at least 10 historical data points
- Check for data quality issues
- Review Prophet error logs

### Performance Issues
- Check DuckDB query plans
- Monitor system resources
- Consider enabling profiling

## License

Part of RTSCS (Real-Time Supply Chain System)

## Support

For issues and feature requests, see the project GitHub repository.
