"""FastAPI routes for analytics service."""
from fastapi import APIRouter, Query, HTTPException
from datetime import datetime
from typing import Optional, List
from app.api.handlers import AnalyticsHandlers, EventHandlers
from app.models.schemas import (
    OrderMetrics,
    InventoryMetrics,
    CustomerMetrics,
    DemandForecast,
    DailyReport,
    HealthResponse
)
from app.analytics.metrics import get_metrics_calculator


router = APIRouter()


# Metrics endpoints
@router.get("/metrics/orders", response_model=OrderMetrics, tags=["Metrics"])
async def get_order_metrics():
    """Get order metrics and statistics.
    
    Returns:
        - total_orders: Count of orders
        - total_revenue: Sum of order amounts
        - average_order_value: Mean order value
        - order_rate_per_day: Orders per day
        - period_start: Start of analysis period
        - period_end: End of analysis period
    """
    return await AnalyticsHandlers.get_order_metrics()


@router.get("/metrics/inventory", response_model=InventoryMetrics, tags=["Metrics"])
async def get_inventory_metrics():
    """Get inventory metrics and statistics.
    
    Returns:
        - total_skus: Count of unique SKUs
        - total_stock_quantity: Sum of all stock
        - low_stock_count: Items below threshold
        - average_turnover_rate: Inventory turnover rate
        - by_sku: Per-SKU inventory details
    """
    return await AnalyticsHandlers.get_inventory_metrics()


@router.get("/metrics/customers", response_model=CustomerMetrics, tags=["Metrics"])
async def get_customer_metrics():
    """Get customer analytics metrics.
    
    Returns:
        - total_customers: Count of unique customers
        - new_customers_this_period: New customers today
        - repeat_customer_count: Repeat customer count
        - repeat_rate: Percentage of repeat customers
        - average_clv: Average customer lifetime value
    """
    return await AnalyticsHandlers.get_customer_metrics()


# Forecasting endpoint
@router.get("/forecast/demand", response_model=DemandForecast, tags=["Forecasting"])
async def get_demand_forecast(
    sku: str = Query(..., description="Product SKU"),
    days: int = Query(30, ge=1, le=365, description="Forecast days (1-365)")
):
    """Generate demand forecast for a SKU.
    
    Uses Prophet time-series forecasting model to predict demand.
    
    Args:
        sku: Product SKU identifier
        days: Number of days to forecast (default: 30)
    
    Returns:
        - sku: Product SKU
        - forecast_days: Number of days forecasted
        - forecast: List of daily forecast points
        - model_type: Forecasting model used (prophet/arima)
        - confidence_level: Model confidence (0-1)
    """
    return await AnalyticsHandlers.get_demand_forecast(sku, days)


# Reports endpoint
@router.get("/reports/daily", response_model=DailyReport, tags=["Reports"])
async def get_daily_report(
    date: Optional[str] = Query(None, description="Report date (YYYY-MM-DD), defaults to today")
):
    """Get comprehensive daily analytics report.
    
    Combines order, inventory, and customer metrics for a single day.
    
    Args:
        date: Report date in YYYY-MM-DD format (optional, defaults to today)
    
    Returns:
        - report_date: Date of the report
        - order_metrics: Daily order statistics
        - inventory_metrics: Daily inventory changes
        - customer_metrics: Daily customer activity
        - generated_at: Report generation timestamp
    """
    return await AnalyticsHandlers.get_daily_report(date)


# Analytics queries
@router.get("/analytics/daily-trends", tags=["Analytics"])
async def get_daily_trends(days: int = Query(30, ge=1, le=365)):
    """Get daily order trends over time period.
    
    Args:
        days: Number of days to include (default: 30)
    
    Returns:
        List of daily metrics with orders, revenue, and averages.
    """
    try:
        calculator = get_metrics_calculator()
        trends = await calculator.get_daily_trends(days)
        return {"trends": trends, "days": days}
    except Exception as e:
        raise HTTPException(status_code=500, detail="Failed to retrieve trends")


@router.get("/analytics/low-stock", tags=["Analytics"])
async def get_low_stock_items():
    """Get items currently at low stock levels.
    
    Returns:
        List of SKUs with current stock below configured threshold.
    """
    try:
        calculator = get_metrics_calculator()
        items = await calculator.get_low_stock_items()
        return {"low_stock_items": items, "count": len(items)}
    except Exception as e:
        raise HTTPException(status_code=500, detail="Failed to retrieve low stock items")


@router.get("/analytics/top-skus", tags=["Analytics"])
async def get_top_revenue_skus(
    limit: int = Query(10, ge=1, le=100),
    days: int = Query(30, ge=1, le=365)
):
    """Get top revenue-generating SKUs.
    
    Args:
        limit: Number of SKUs to return (default: 10)
        days: Analysis period in days (default: 30)
    
    Returns:
        List of SKUs ranked by average revenue per order.
    """
    try:
        calculator = get_metrics_calculator()
        skus = await calculator.get_top_revenue_skus(limit, days)
        return {"top_skus": skus, "period_days": days}
    except Exception as e:
        raise HTTPException(status_code=500, detail="Failed to retrieve top SKUs")


@router.get("/analytics/customer-value", tags=["Analytics"])
async def get_customer_lifetime_value(
    limit: int = Query(10, ge=1, le=100)
):
    """Get customers by lifetime value.
    
    Args:
        limit: Number of customers to return (default: 10)
    
    Returns:
        List of customers ranked by total lifetime spend.
    """
    try:
        calculator = get_metrics_calculator()
        customers = await calculator.get_customer_lifetime_value(limit)
        return {"customers": customers, "count": len(customers)}
    except Exception as e:
        raise HTTPException(status_code=500, detail="Failed to retrieve customer value metrics")


# Health check
@router.get("/health", response_model=HealthResponse, tags=["Health"])
async def health_check():
    """Health check endpoint.
    
    Returns:
        Service health status and dependency availability.
    """
    return await AnalyticsHandlers.health_check()


# Root endpoint
@router.get("/", tags=["Info"])
async def root():
    """Analytics Service root endpoint."""
    return {
        "service": "Analytics Service (Phase 2)",
        "version": "2.0.0",
        "endpoints": {
            "metrics": ["/metrics/orders", "/metrics/inventory", "/metrics/customers"],
            "forecasting": ["/forecast/demand"],
            "reports": ["/reports/daily"],
            "analytics": ["/analytics/daily-trends", "/analytics/low-stock", "/analytics/top-skus", "/analytics/customer-value"],
            "health": ["/health"]
        }
    }
