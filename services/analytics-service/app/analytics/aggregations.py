"""Aggregations and batch processing."""
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional
import pandas as pd
from app.database.duckdb_client import get_db_client
from app.database.queries import UPDATE_CUSTOMER_METRICS
from config import settings


logger = logging.getLogger(__name__)


class AggregationEngine:
    """Engine for aggregating metrics and generating reports."""
    
    def __init__(self):
        """Initialize aggregation engine."""
        self.db_client = get_db_client()
    
    async def aggregate_hourly_metrics(self) -> Dict[str, Any]:
        """Aggregate metrics for the last hour."""
        try:
            now = datetime.now()
            hour_ago = now - timedelta(hours=1)
            
            metrics = {
                "period": "hourly",
                "start_time": hour_ago,
                "end_time": now,
                "generated_at": now
            }
            
            # Get order metrics
            order_result = self.db_client.fetch_one("""
                SELECT
                    COUNT(DISTINCT order_id) as order_count,
                    SUM(total_amount) as total_revenue
                FROM orders_fact
                WHERE created_at >= ? AND created_at < ?
            """, [hour_ago, now])
            
            metrics["orders"] = {
                "count": order_result.get("order_count", 0) if order_result else 0,
                "revenue": float(order_result.get("total_revenue", 0) or 0) if order_result else 0.0
            }
            
            # Get inventory updates
            inv_result = self.db_client.fetch_one("""
                SELECT
                    COUNT(*) as update_count,
                    COUNT(DISTINCT sku) as sku_count
                FROM inventory_fact
                WHERE updated_at >= ? AND updated_at < ?
            """, [hour_ago, now])
            
            metrics["inventory"] = {
                "updates": inv_result.get("update_count", 0) if inv_result else 0,
                "skus_updated": inv_result.get("sku_count", 0) if inv_result else 0
            }
            
            return metrics
        
        except Exception as e:
            logger.error(f"Failed to aggregate hourly metrics: {e}")
            raise
    
    async def aggregate_daily_metrics(self, date: Optional[datetime] = None) -> Dict[str, Any]:
        """Aggregate metrics for a specific day."""
        try:
            if date is None:
                date = datetime.now()
            
            day_start = datetime.combine(date.date(), datetime.min.time())
            day_end = day_start + timedelta(days=1)
            
            metrics = {
                "date": str(date.date()),
                "generated_at": datetime.now()
            }
            
            # Order metrics
            order_result = self.db_client.fetch_one("""
                SELECT
                    COUNT(DISTINCT order_id) as order_count,
                    SUM(total_amount) as total_revenue,
                    AVG(total_amount) as avg_order_value,
                    COUNT(DISTINCT customer_id) as unique_customers
                FROM orders_fact
                WHERE created_at >= ? AND created_at < ? AND status = 'created'
            """, [day_start, day_end])
            
            metrics["orders"] = {
                "count": order_result.get("order_count", 0) if order_result else 0,
                "revenue": float(order_result.get("total_revenue", 0) or 0) if order_result else 0.0,
                "avg_value": float(order_result.get("avg_order_value", 0) or 0) if order_result else 0.0,
                "unique_customers": order_result.get("unique_customers", 0) if order_result else 0
            }
            
            # Cancelled orders
            cancelled_result = self.db_client.fetch_one("""
                SELECT COUNT(*) as cancelled_count
                FROM orders_fact
                WHERE cancelled_at >= ? AND cancelled_at < ?
            """, [day_start, day_end])
            
            metrics["orders"]["cancelled"] = cancelled_result.get("cancelled_count", 0) if cancelled_result else 0
            
            # Inventory changes
            inv_result = self.db_client.fetch_one("""
                SELECT
                    COUNT(*) as update_count,
                    COUNT(DISTINCT sku) as sku_count,
                    SUM(CASE WHEN is_low_stock THEN 1 ELSE 0 END) as low_stock_alerts
                FROM inventory_fact
                WHERE updated_at >= ? AND updated_at < ?
            """, [day_start, day_end])
            
            metrics["inventory"] = {
                "updates": inv_result.get("update_count", 0) if inv_result else 0,
                "skus_changed": inv_result.get("sku_count", 0) if inv_result else 0,
                "low_stock_alerts": inv_result.get("low_stock_alerts", 0) if inv_result else 0
            }
            
            # Customer metrics
            cust_result = self.db_client.fetch_one("""
                SELECT
                    COUNT(DISTINCT customer_id) as new_customers
                FROM customers_dim
                WHERE DATE(first_seen) = ?
            """, [date.date()])
            
            metrics["customers"] = {
                "new": cust_result.get("new_customers", 0) if cust_result else 0
            }
            
            return metrics
        
        except Exception as e:
            logger.error(f"Failed to aggregate daily metrics: {e}")
            raise
    
    async def update_customer_dimensions(self) -> int:
        """Update customer dimension table from orders."""
        try:
            self.db_client.execute(UPDATE_CUSTOMER_METRICS)
            logger.info("Customer dimensions updated")
            return 1
        except Exception as e:
            logger.error(f"Failed to update customer dimensions: {e}")
            raise
    
    async def export_daily_parquet(self, date: Optional[datetime] = None) -> Dict[str, str]:
        """Export daily data to parquet files."""
        try:
            if date is None:
                date = datetime.now()
            
            day_start = datetime.combine(date.date(), datetime.min.time())
            day_end = day_start + timedelta(days=1)
            date_str = date.strftime('%Y%m%d')
            
            paths = {}
            
            # Export orders
            orders_query = f"""
                SELECT * FROM orders_fact
                WHERE created_at >= '{day_start}' AND created_at < '{day_end}'
            """
            orders_df = self.db_client.fetch_df(orders_query)
            
            orders_path = f"{settings.DB_PARQUET_PATH}/orders_{date_str}.parquet"
            orders_df.to_parquet(orders_path)
            paths["orders"] = orders_path
            logger.info(f"Exported {len(orders_df)} orders to {orders_path}")
            
            # Export inventory
            inv_query = f"""
                SELECT * FROM inventory_fact
                WHERE updated_at >= '{day_start}' AND updated_at < '{day_end}'
            """
            inv_df = self.db_client.fetch_df(inv_query)
            
            inv_path = f"{settings.DB_PARQUET_PATH}/inventory_{date_str}.parquet"
            inv_df.to_parquet(inv_path)
            paths["inventory"] = inv_path
            logger.info(f"Exported {len(inv_df)} inventory records to {inv_path}")
            
            return paths
        
        except Exception as e:
            logger.error(f"Failed to export parquet data: {e}")
            raise
    
    async def cleanup_old_data(self, retention_days: int = 90) -> int:
        """Clean up old data beyond retention period."""
        try:
            cutoff_date = datetime.now() - timedelta(days=retention_days)
            
            # Delete old forecast data
            deleted = self.db_client.execute(f"""
                DELETE FROM forecasts
                WHERE generated_at < '{cutoff_date}'
            """)
            
            logger.info(f"Cleaned up data older than {retention_days} days")
            return 1
        except Exception as e:
            logger.error(f"Failed to cleanup old data: {e}")
            raise


# Global instance
_aggregation_engine: Optional[AggregationEngine] = None


def get_aggregation_engine() -> AggregationEngine:
    """Get aggregation engine instance."""
    global _aggregation_engine
    if _aggregation_engine is None:
        _aggregation_engine = AggregationEngine()
    return _aggregation_engine
