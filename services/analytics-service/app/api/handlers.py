"""API route handlers."""
import logging
from datetime import datetime
from typing import Dict, Any, Optional
from fastapi import HTTPException, Query
from app.models.schemas import (
    OrderMetrics,
    InventoryMetrics,
    CustomerMetrics,
    DemandForecast,
    DailyReport,
    HealthResponse,
    OrderCreatedEvent,
    OrderCancelledEvent,
    StockUpdatedEvent,
    LowStockAlertEvent
)
from app.analytics.metrics import get_metrics_calculator
from app.analytics.forecasting import get_forecaster
from app.analytics.aggregations import get_aggregation_engine
from app.database.duckdb_client import get_db_client
from app.kafka.producer import get_kafka_producer
from config import settings


logger = logging.getLogger(__name__)


class AnalyticsHandlers:
    """Handlers for analytics endpoints."""
    
    @staticmethod
    async def get_order_metrics() -> OrderMetrics:
        """Get order metrics."""
        try:
            calculator = get_metrics_calculator()
            metrics = await calculator.get_order_metrics()
            return OrderMetrics(**metrics)
        except Exception as e:
            logger.error(f"Failed to get order metrics: {e}")
            raise HTTPException(status_code=500, detail="Failed to retrieve order metrics")
    
    @staticmethod
    async def get_inventory_metrics() -> InventoryMetrics:
        """Get inventory metrics."""
        try:
            calculator = get_metrics_calculator()
            metrics = await calculator.get_inventory_metrics()
            return InventoryMetrics(**metrics)
        except Exception as e:
            logger.error(f"Failed to get inventory metrics: {e}")
            raise HTTPException(status_code=500, detail="Failed to retrieve inventory metrics")
    
    @staticmethod
    async def get_customer_metrics() -> CustomerMetrics:
        """Get customer metrics."""
        try:
            calculator = get_metrics_calculator()
            metrics = await calculator.get_customer_metrics()
            return CustomerMetrics(**metrics)
        except Exception as e:
            logger.error(f"Failed to get customer metrics: {e}")
            raise HTTPException(status_code=500, detail="Failed to retrieve customer metrics")
    
    @staticmethod
    async def get_demand_forecast(
        sku: str,
        days: int = Query(30, ge=1, le=365)
    ) -> DemandForecast:
        """Get demand forecast for SKU."""
        try:
            if not sku:
                raise HTTPException(status_code=400, detail="SKU is required")
            
            forecaster = get_forecaster()
            forecast = await forecaster.forecast_demand(sku, days, "prophet")
            
            if not forecast:
                raise HTTPException(status_code=404, detail=f"Could not generate forecast for {sku}")
            
            # Publish forecast event
            producer = get_kafka_producer()
            await producer.publish_forecast_event(forecast)
            
            return DemandForecast(**forecast)
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Failed to get forecast for {sku}: {e}")
            raise HTTPException(status_code=500, detail="Failed to generate forecast")
    
    @staticmethod
    async def get_daily_report(date: Optional[str] = None) -> DailyReport:
        """Get daily analytics report."""
        try:
            if date:
                try:
                    from datetime import datetime
                    report_date = datetime.strptime(date, "%Y-%m-%d")
                except ValueError:
                    raise HTTPException(status_code=400, detail="Invalid date format, use YYYY-MM-DD")
            else:
                report_date = datetime.now()
            
            aggregation_engine = get_aggregation_engine()
            metrics_calculator = get_metrics_calculator()
            
            # Get all metrics
            order_metrics = await metrics_calculator.get_order_metrics()
            inventory_metrics = await metrics_calculator.get_inventory_metrics()
            customer_metrics = await metrics_calculator.get_customer_metrics()
            
            return DailyReport(
                report_date=str(report_date.date()),
                order_metrics=OrderMetrics(**order_metrics),
                inventory_metrics=InventoryMetrics(**inventory_metrics),
                customer_metrics=CustomerMetrics(**customer_metrics),
                generated_at=datetime.now()
            )
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Failed to get daily report: {e}")
            raise HTTPException(status_code=500, detail="Failed to generate report")
    
    @staticmethod
    async def health_check() -> HealthResponse:
        """Health check endpoint."""
        try:
            db_client = get_db_client()
            db_connected = db_client.table_exists("orders_fact")
            
            producer = get_kafka_producer()
            kafka_connected = False
            try:
                if not producer.connection:
                    await producer.connect()
                kafka_connected = producer.connection is not None
            except Exception:
                kafka_connected = False
            
            return HealthResponse(
                status="healthy" if db_connected and kafka_connected else "degraded",
                service=settings.SERVICE_NAME,
                version="2.0.0",
                database_connected=db_connected,
                kafka_connected=kafka_connected,
                timestamp=datetime.now()
            )
        except Exception as e:
            logger.error(f"Health check failed: {e}")
            return HealthResponse(
                status="unhealthy",
                service=settings.SERVICE_NAME,
                version="2.0.0",
                database_connected=False,
                kafka_connected=False,
                timestamp=datetime.now()
            )


class EventHandlers:
    """Handlers for event processing."""
    
    @staticmethod
    async def handle_order_created(event: OrderCreatedEvent) -> None:
        """Handle order created event."""
        try:
            db_client = get_db_client()
            
            # Convert SKU list to array
            skus = event.sku_list if event.sku_list else []
            
            data = [{
                "order_id": event.order_id,
                "customer_id": event.customer_id,
                "total_amount": event.total_amount,
                "item_count": event.item_count,
                "status": "created",
                "created_at": event.timestamp,
                "sku_list": skus,
                "event_id": event.event_id
            }]
            
            db_client.insert_data("orders_fact", data)
            logger.info(f"Inserted order: {event.order_id}")
        
        except Exception as e:
            logger.error(f"Failed to handle order created event: {e}")
    
    @staticmethod
    async def handle_order_cancelled(event: OrderCancelledEvent) -> None:
        """Handle order cancelled event."""
        try:
            db_client = get_db_client()
            
            query = """
                UPDATE orders_fact
                SET status = 'cancelled', cancelled_at = ?
                WHERE order_id = ?
            """
            
            db_client.execute(query, [event.timestamp, event.order_id])
            logger.info(f"Cancelled order: {event.order_id}")
        
        except Exception as e:
            logger.error(f"Failed to handle order cancelled event: {e}")
    
    @staticmethod
    async def handle_stock_updated(event: StockUpdatedEvent) -> None:
        """Handle stock updated event."""
        try:
            db_client = get_db_client()
            
            data = [{
                "sku": event.sku,
                "warehouse_id": event.warehouse_id,
                "previous_quantity": event.previous_quantity,
                "new_quantity": event.new_quantity,
                "updated_at": event.timestamp,
                "is_low_stock": False,
                "event_id": event.event_id
            }]
            
            db_client.insert_data("inventory_fact", data)
            logger.info(f"Updated stock for {event.sku}: {event.new_quantity}")
        
        except Exception as e:
            logger.error(f"Failed to handle stock updated event: {e}")
    
    @staticmethod
    async def handle_low_stock_alert(event: LowStockAlertEvent) -> None:
        """Handle low stock alert event."""
        try:
            db_client = get_db_client()
            
            data = [{
                "sku": event.sku,
                "warehouse_id": None,
                "previous_quantity": None,
                "new_quantity": event.current_quantity,
                "updated_at": event.timestamp,
                "alert_threshold": event.threshold,
                "is_low_stock": True,
                "event_id": event.event_id
            }]
            
            db_client.insert_data("inventory_fact", data)
            logger.warning(f"Low stock alert for {event.sku}: {event.current_quantity}")
        
        except Exception as e:
            logger.error(f"Failed to handle low stock alert event: {e}")
