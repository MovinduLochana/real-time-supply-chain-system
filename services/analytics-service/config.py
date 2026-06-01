"""Configuration module for Analytics Service."""
from pydantic_settings import BaseSettings
from typing import Optional
import logging


class Settings(BaseSettings):
    """Application settings."""
    
    # Service
    SERVICE_NAME: str = "analytics-service"
    SERVICE_PORT: int = 3003
    SERVICE_HOST: str = "0.0.0.0"
    DEBUG: bool = False
    
    # Kafka
    KAFKA_BROKERS: str = "localhost:9092"
    KAFKA_ORDER_EVENTS_TOPIC: str = "order-events"
    KAFKA_INVENTORY_EVENTS_TOPIC: str = "inventory-events"
    KAFKA_METRICS_TOPIC: str = "metrics-events"
    KAFKA_FORECAST_TOPIC: str = "forecast-events"
    KAFKA_CONSUMER_GROUP: str = "analytics-service"
    KAFKA_AUTO_OFFSET_RESET: str = "earliest"
    KAFKA_SESSION_TIMEOUT_MS: int = 30000
    KAFKA_HEARTBEAT_INTERVAL_MS: int = 10000
    
    # Database
    DB_PATH: str = "./data/analytics.duckdb"
    DB_PARQUET_PATH: str = "./data/parquet"
    
    # OpenTelemetry
    JAEGER_HOST: str = "localhost"
    JAEGER_PORT: int = 6831
    OTEL_ENABLED: bool = False
    
    # Logging
    LOG_LEVEL: str = "INFO"
    LOG_FORMAT: str = "json"
    
    # Analytics
    METRICS_CACHE_TTL_SECONDS: int = 300
    FORECAST_DAYS_DEFAULT: int = 30
    FORECAST_DAYS_MAX: int = 365
    DAILY_REPORT_HOUR: int = 0
    DAILY_REPORT_MINUTE: int = 0
    
    # Performance
    CONSUMER_BATCH_SIZE: int = 100
    CONSUMER_BATCH_TIMEOUT_MS: int = 5000
    
    class Config:
        env_file = ".env"
        case_sensitive = True


settings = Settings()


def get_logger(name: str, level: str = None) -> logging.Logger:
    """Get configured logger instance."""
    logger = logging.getLogger(name)
    if level is None:
        level = settings.LOG_LEVEL
    logger.setLevel(getattr(logging, level.upper()))
    return logger
