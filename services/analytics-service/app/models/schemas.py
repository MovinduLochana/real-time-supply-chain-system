"""Data models and schemas."""
from pydantic import BaseModel, Field
from typing import Optional, Dict, Any, List
from datetime import datetime
from enum import Enum


class EventType(str, Enum):
    """Event types."""
    ORDER_CREATED = "order_created"
    ORDER_CANCELLED = "order_cancelled"
    STOCK_UPDATED = "stock_updated"
    LOW_STOCK_ALERT = "low_stock_alert"
    METRICS_AGGREGATED = "metrics_aggregated"
    FORECAST_GENERATED = "forecast_generated"


class OrderCreatedEvent(BaseModel):
    """Order created event."""
    event_id: str
    order_id: str
    customer_id: str
    total_amount: float
    item_count: int
    timestamp: datetime
    sku_list: List[str]


class OrderCancelledEvent(BaseModel):
    """Order cancelled event."""
    event_id: str
    order_id: str
    timestamp: datetime
    reason: Optional[str] = None


class StockUpdatedEvent(BaseModel):
    """Stock updated event."""
    event_id: str
    sku: str
    previous_quantity: int
    new_quantity: int
    timestamp: datetime
    warehouse_id: Optional[str] = None


class LowStockAlertEvent(BaseModel):
    """Low stock alert event."""
    event_id: str
    sku: str
    current_quantity: int
    threshold: int
    timestamp: datetime


class MetricsAggregatedEvent(BaseModel):
    """Metrics aggregated event."""
    event_id: str
    timestamp: datetime
    period: str  # hourly, daily, weekly
    metrics: Dict[str, Any]


class ForecastGeneratedEvent(BaseModel):
    """Forecast generated event."""
    event_id: str
    sku: str
    days: int
    forecast_data: List[Dict[str, Any]]
    model_type: str  # prophet, arima
    timestamp: datetime


# API Response Models
class OrderMetrics(BaseModel):
    """Order metrics response."""
    total_orders: int
    total_revenue: float
    average_order_value: float
    order_rate_per_day: float
    period_start: datetime
    period_end: datetime
    timestamp: datetime


class InventoryMetrics(BaseModel):
    """Inventory metrics response."""
    total_skus: int
    total_stock_quantity: int
    low_stock_count: int
    average_turnover_rate: float
    by_sku: Dict[str, Dict[str, Any]]
    timestamp: datetime


class CustomerMetrics(BaseModel):
    """Customer metrics response."""
    total_customers: int
    new_customers_this_period: int
    repeat_customer_count: int
    repeat_rate: float
    average_clv: float
    timestamp: datetime


class DemandForecast(BaseModel):
    """Demand forecast response."""
    sku: str
    forecast_days: int
    forecast: List[Dict[str, Any]]
    model_type: str
    confidence_level: float
    generated_at: datetime


class DailyReport(BaseModel):
    """Daily report response."""
    report_date: str
    order_metrics: OrderMetrics
    inventory_metrics: InventoryMetrics
    customer_metrics: CustomerMetrics
    generated_at: datetime


class HealthResponse(BaseModel):
    """Health check response."""
    status: str
    service: str
    version: str
    database_connected: bool
    kafka_connected: bool
    timestamp: datetime
