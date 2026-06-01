"""Analytics metrics calculations."""
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional
import pandas as pd
from app.database.duckdb_client import get_db_client
from app.database.queries import (
    GET_ORDER_METRICS,
    GET_ORDER_METRICS_BY_DATE,
    GET_INVENTORY_METRICS,
    GET_INVENTORY_BY_SKU,
    GET_CUSTOMER_METRICS,
    GET_LOW_STOCK_ITEMS,
    GET_TOP_REVENUE_SKUS,
    GET_CUSTOMER_LIFETIME_VALUE
)


logger = logging.getLogger(__name__)


class MetricsCalculator:
    """Calculate analytics metrics."""
    
    def __init__(self):
        """Initialize metrics calculator."""
        self.db_client = get_db_client()
    
    async def get_order_metrics(self) -> Dict[str, Any]:
        """Get order metrics."""
        try:
            result = self.db_client.fetch_one(GET_ORDER_METRICS)
            if not result:
                return {
                    "total_orders": 0,
                    "total_revenue": 0.0,
                    "average_order_value": 0.0,
                    "order_rate_per_day": 0.0,
                    "period_start": datetime.now(),
                    "period_end": datetime.now(),
                    "timestamp": datetime.now()
                }
            
            return {
                "total_orders": result.get("total_orders", 0),
                "total_revenue": float(result.get("total_revenue", 0.0) or 0.0),
                "average_order_value": float(result.get("avg_order_value", 0.0) or 0.0),
                "order_rate_per_day": float(result.get("order_rate_per_day", 0.0) or 0.0),
                "period_start": result.get("period_start", datetime.now()),
                "period_end": result.get("period_end", datetime.now()),
                "timestamp": datetime.now()
            }
        except Exception as e:
            logger.error(f"Failed to get order metrics: {e}")
            raise
    
    async def get_inventory_metrics(self) -> Dict[str, Any]:
        """Get inventory metrics."""
        try:
            result = self.db_client.fetch_one(GET_INVENTORY_METRICS)
            
            if not result:
                return {
                    "total_skus": 0,
                    "total_stock_quantity": 0,
                    "low_stock_count": 0,
                    "average_turnover_rate": 0.0,
                    "by_sku": {},
                    "timestamp": datetime.now()
                }
            
            # Get detailed by-SKU metrics
            by_sku_results = self.db_client.fetch_all(
                GET_INVENTORY_BY_SKU,
                [datetime.now() - timedelta(days=30), datetime.now()]
            )
            
            by_sku = {}
            for row in by_sku_results:
                by_sku[row.get("sku")] = {
                    "current_stock": row.get("new_quantity", 0),
                    "net_change": row.get("net_change", 0),
                    "update_count": row.get("update_count", 0),
                    "last_updated": row.get("last_updated")
                }
            
            return {
                "total_skus": result.get("total_skus", 0),
                "total_stock_quantity": result.get("total_stock_quantity", 0),
                "low_stock_count": result.get("low_stock_count", 0),
                "average_turnover_rate": float(result.get("avg_stock_level", 0.0) or 0.0),
                "by_sku": by_sku,
                "timestamp": datetime.now()
            }
        except Exception as e:
            logger.error(f"Failed to get inventory metrics: {e}")
            raise
    
    async def get_customer_metrics(self) -> Dict[str, Any]:
        """Get customer metrics."""
        try:
            result = self.db_client.fetch_one(GET_CUSTOMER_METRICS)
            
            if not result:
                return {
                    "total_customers": 0,
                    "new_customers_this_period": 0,
                    "repeat_customer_count": 0,
                    "repeat_rate": 0.0,
                    "average_clv": 0.0,
                    "timestamp": datetime.now()
                }
            
            return {
                "total_customers": result.get("total_customers", 0),
                "new_customers_this_period": result.get("new_customers_this_period", 0),
                "repeat_customer_count": result.get("repeat_customer_count", 0),
                "repeat_rate": float(result.get("repeat_rate", 0.0) or 0.0),
                "average_clv": float(result.get("avg_clv", 0.0) or 0.0),
                "timestamp": datetime.now()
            }
        except Exception as e:
            logger.error(f"Failed to get customer metrics: {e}")
            raise
    
    async def get_low_stock_items(self) -> List[Dict[str, Any]]:
        """Get items with low stock."""
        try:
            results = self.db_client.fetch_all(GET_LOW_STOCK_ITEMS)
            return [
                {
                    "sku": row.get("sku"),
                    "current_stock": row.get("current_stock", 0),
                    "alert_threshold": row.get("alert_threshold"),
                    "last_alert": row.get("last_alert")
                }
                for row in results
            ]
        except Exception as e:
            logger.error(f"Failed to get low stock items: {e}")
            return []
    
    async def get_top_revenue_skus(self, limit: int = 10, days: int = 30) -> List[Dict[str, Any]]:
        """Get top revenue generating SKUs."""
        try:
            start_date = datetime.now() - timedelta(days=days)
            results = self.db_client.fetch_all(
                GET_TOP_REVENUE_SKUS,
                [start_date, limit]
            )
            return [
                {
                    "sku": row.get("sku"),
                    "order_count": row.get("order_count", 0),
                    "avg_revenue_per_order": float(row.get("avg_revenue_per_order", 0.0) or 0.0)
                }
                for row in results
            ]
        except Exception as e:
            logger.error(f"Failed to get top revenue SKUs: {e}")
            return []
    
    async def get_customer_lifetime_value(self, limit: int = 10) -> List[Dict[str, Any]]:
        """Get customer lifetime value metrics."""
        try:
            results = self.db_client.fetch_all(GET_CUSTOMER_LIFETIME_VALUE, [limit])
            return [
                {
                    "customer_id": row.get("customer_id"),
                    "total_orders": row.get("total_orders", 0),
                    "total_spent": float(row.get("total_spent", 0.0) or 0.0),
                    "avg_order_value": float(row.get("avg_order_value", 0.0) or 0.0),
                    "customer_tenure_days": row.get("customer_tenure_days", 0)
                }
                for row in results
            ]
        except Exception as e:
            logger.error(f"Failed to get CLV metrics: {e}")
            return []
    
    async def get_daily_trends(self, days: int = 30) -> List[Dict[str, Any]]:
        """Get daily order trends."""
        try:
            end_date = datetime.now()
            start_date = end_date - timedelta(days=days)
            
            results = self.db_client.fetch_all(
                GET_ORDER_METRICS_BY_DATE,
                [start_date, end_date]
            )
            
            return [
                {
                    "date": str(row.get("date")),
                    "orders": row.get("orders", 0),
                    "revenue": float(row.get("revenue", 0.0) or 0.0),
                    "avg_order_value": float(row.get("avg_order_value", 0.0) or 0.0)
                }
                for row in results
            ]
        except Exception as e:
            logger.error(f"Failed to get daily trends: {e}")
            return []


# Global instance
_metrics_calculator: Optional[MetricsCalculator] = None


def get_metrics_calculator() -> MetricsCalculator:
    """Get metrics calculator instance."""
    global _metrics_calculator
    if _metrics_calculator is None:
        _metrics_calculator = MetricsCalculator()
    return _metrics_calculator
