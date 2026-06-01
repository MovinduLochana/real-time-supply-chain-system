"""Forecasting module for demand prediction."""
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional
import pandas as pd
import numpy as np
from app.database.duckdb_client import get_db_client
from config import settings


logger = logging.getLogger(__name__)


class DemandForecaster:
    """Forecast demand using time-series models."""
    
    def __init__(self):
        """Initialize forecaster."""
        self.db_client = get_db_client()
        self.min_data_points = 10
    
    async def forecast_demand(
        self,
        sku: str,
        days: int = 30,
        model_type: str = "prophet"
    ) -> Optional[Dict[str, Any]]:
        """Forecast demand for a SKU."""
        try:
            # Validate input
            days = min(days, settings.FORECAST_DAYS_MAX)
            if days < 1:
                days = settings.FORECAST_DAYS_DEFAULT
            
            # Get historical order data for SKU
            historical_data = await self._get_historical_data(sku)
            
            if len(historical_data) < self.min_data_points:
                logger.warning(f"Insufficient data for {sku}: {len(historical_data)} points")
                return None
            
            # Prepare time series
            ts_data = pd.DataFrame(historical_data)
            ts_data['date'] = pd.to_datetime(ts_data['date'])
            ts_data = ts_data.sort_values('date')
            
            # Generate forecast
            if model_type.lower() == "prophet":
                forecast = await self._forecast_with_prophet(ts_data, sku, days)
            else:
                forecast = await self._forecast_with_arima(ts_data, sku, days)
            
            return forecast
        
        except Exception as e:
            logger.error(f"Forecasting failed for {sku}: {e}")
            return None
    
    async def _get_historical_data(self, sku: str) -> List[Dict[str, Any]]:
        """Get historical order data for SKU."""
        try:
            query = """
                SELECT
                    DATE(created_at) as date,
                    COUNT(*) as order_count,
                    SUM(total_amount) as revenue
                FROM orders_fact, UNNEST(sku_list) as sku_item(value)
                WHERE sku_item.value = ?
                AND status = 'created'
                AND created_at >= CURRENT_DATE - INTERVAL 365 DAY
                GROUP BY DATE(created_at)
                ORDER BY date
            """
            
            results = self.db_client.fetch_all(query, [sku])
            return [
                {
                    "date": row.get("date"),
                    "y": float(row.get("order_count", 0))
                }
                for row in results
            ]
        except Exception as e:
            logger.error(f"Failed to get historical data for {sku}: {e}")
            return []
    
    async def _forecast_with_prophet(
        self,
        data: pd.DataFrame,
        sku: str,
        periods: int
    ) -> Dict[str, Any]:
        """Forecast using Prophet model."""
        try:
            from prophet import Prophet
            
            # Prepare data for Prophet
            prophet_data = data[['date', 'y']].copy()
            prophet_data.columns = ['ds', 'y']
            
            # Handle missing values
            prophet_data['y'] = prophet_data['y'].fillna(0)
            
            # Fit model
            model = Prophet(
                interval_width=0.95,
                daily_seasonality=False,
                weekly_seasonality=True,
                yearly_seasonality=True,
                growth='linear'
            )
            model.fit(prophet_data)
            
            # Generate forecast
            future = model.make_future_dataframe(periods=periods)
            forecast = model.predict(future)
            
            # Extract relevant columns
            forecast_data = []
            for _, row in forecast[forecast['ds'] > data['date'].max()].iterrows():
                forecast_data.append({
                    "date": row['ds'].strftime('%Y-%m-%d'),
                    "yhat": float(row['yhat']),
                    "yhat_lower": float(row['yhat_lower']),
                    "yhat_upper": float(row['yhat_upper']),
                    "trend": float(row['trend'])
                })
            
            # Calculate confidence
            confidence = float(np.mean([
                abs(forecast_data[i]['yhat'] - forecast_data[i]['yhat_lower']) / 
                max(abs(forecast_data[i]['yhat']), 1)
                for i in range(len(forecast_data))
            ]))
            confidence = min(0.95, confidence)
            
            return {
                "sku": sku,
                "forecast_days": periods,
                "forecast": forecast_data,
                "model_type": "prophet",
                "confidence_level": float(max(0.7, 1.0 - confidence)),
                "generated_at": datetime.now()
            }
        
        except ImportError:
            logger.warning("Prophet not available, falling back to ARIMA")
            return await self._forecast_with_arima(data, sku, periods)
        except Exception as e:
            logger.error(f"Prophet forecasting failed for {sku}: {e}")
            return None
    
    async def _forecast_with_arima(
        self,
        data: pd.DataFrame,
        sku: str,
        periods: int
    ) -> Dict[str, Any]:
        """Forecast using ARIMA model."""
        try:
            from statsmodels.tsa.arima.model import ARIMA
            
            # Prepare data
            ts = pd.Series(data['y'].values, index=data['date'])
            
            # Auto-detect ARIMA parameters (simple approach)
            # In production, use auto_arima or grid search
            try:
                model = ARIMA(ts, order=(1, 1, 1))
                fitted_model = model.fit()
            except Exception:
                # Fallback to simpler model
                model = ARIMA(ts, order=(0, 1, 0))
                fitted_model = model.fit()
            
            # Generate forecast
            forecast_result = fitted_model.get_forecast(steps=periods)
            forecast_df = forecast_result.summary_frame()
            
            # Extract forecast data
            forecast_data = []
            for idx, (_, row) in enumerate(forecast_df.iterrows()):
                forecast_data.append({
                    "date": (data['date'].max() + timedelta(days=idx+1)).strftime('%Y-%m-%d'),
                    "yhat": float(row['mean']),
                    "yhat_lower": float(row['mean_ci_lower']),
                    "yhat_upper": float(row['mean_ci_upper']),
                    "trend": float(row['mean'])
                })
            
            return {
                "sku": sku,
                "forecast_days": periods,
                "forecast": forecast_data,
                "model_type": "arima",
                "confidence_level": 0.90,
                "generated_at": datetime.now()
            }
        
        except Exception as e:
            logger.error(f"ARIMA forecasting failed for {sku}: {e}")
            return None
    
    async def forecast_multiple_skus(
        self,
        skus: List[str],
        days: int = 30
    ) -> List[Dict[str, Any]]:
        """Forecast demand for multiple SKUs."""
        forecasts = []
        for sku in skus:
            forecast = await self.forecast_demand(sku, days)
            if forecast:
                forecasts.append(forecast)
        return forecasts


# Global instance
_forecaster: Optional[DemandForecaster] = None


def get_forecaster() -> DemandForecaster:
    """Get forecaster instance."""
    global _forecaster
    if _forecaster is None:
        _forecaster = DemandForecaster()
    return _forecaster
