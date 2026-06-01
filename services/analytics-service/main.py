"""Main application entry point."""
import asyncio
import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from datetime import datetime

from config import settings
from app.utils.logging import setup_logging, get_logger
from app.utils.tracing import setup_tracing, setup_metrics
from app.database.duckdb_client import init_db, close_db, get_db_client
from app.kafka.consumer import get_kafka_consumer
from app.kafka.producer import get_kafka_producer
from app.api.handlers import EventHandlers
from app.api.routes import router


# Setup logging
setup_logging()
logger = get_logger(__name__)

# Setup tracing
tracer = setup_tracing()
meter = setup_metrics()


# Kafka consumer task
kafka_consumer_task = None


async def start_kafka_consumer():
    """Start consuming Kafka events."""
    try:
        consumer = get_kafka_consumer()
        
        # Register event handlers
        consumer.register_handler("OrderCreatedEvent", EventHandlers.handle_order_created)
        consumer.register_handler("OrderCancelledEvent", EventHandlers.handle_order_cancelled)
        consumer.register_handler("StockUpdatedEvent", EventHandlers.handle_stock_updated)
        consumer.register_handler("LowStockAlertEvent", EventHandlers.handle_low_stock_alert)
        
        # Start consumers
        await consumer.connect()
        
        # Run both consumers concurrently
        await asyncio.gather(
            consumer.consume_order_events(),
            consumer.consume_inventory_events(),
            return_exceptions=True
        )
    
    except Exception as e:
        logger.error(f"Kafka consumer error: {e}")


async def startup_event():
    """FastAPI startup event."""
    global kafka_consumer_task
    
    logger.info(f"Starting {settings.SERVICE_NAME} v2.0.0")
    
    # Initialize database
    await init_db()
    
    # Connect producer
    try:
        producer = get_kafka_producer()
        await producer.connect()
        logger.info("Kafka producer connected")
    except Exception as e:
        logger.warning(f"Kafka producer connection failed: {e}")
    
    # Start Kafka consumer as background task
    try:
        kafka_consumer_task = asyncio.create_task(start_kafka_consumer())
        logger.info("Kafka consumer started")
    except Exception as e:
        logger.warning(f"Failed to start Kafka consumer: {e}")


async def shutdown_event():
    """FastAPI shutdown event."""
    global kafka_consumer_task
    
    logger.info(f"Shutting down {settings.SERVICE_NAME}")
    
    # Stop Kafka consumer
    try:
        consumer = get_kafka_consumer()
        await consumer.disconnect()
        if kafka_consumer_task:
            kafka_consumer_task.cancel()
        logger.info("Kafka consumer stopped")
    except Exception as e:
        logger.warning(f"Error stopping Kafka consumer: {e}")
    
    # Disconnect producer
    try:
        producer = get_kafka_producer()
        await producer.disconnect()
        logger.info("Kafka producer disconnected")
    except Exception as e:
        logger.warning(f"Error disconnecting producer: {e}")
    
    # Close database
    await close_db()
    logger.info("Database closed")


@asynccontextmanager
async def lifespan(app: FastAPI):
    """FastAPI lifespan context manager."""
    # Startup
    await startup_event()
    yield
    # Shutdown
    await shutdown_event()


# Create FastAPI app
app = FastAPI(
    title="Analytics Service",
    description="Real-time analytics and forecasting service for RTSCS",
    version="2.0.0",
    lifespan=lifespan
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routes
app.include_router(router)


@app.get("/info", tags=["Info"])
async def service_info():
    """Get service information."""
    return {
        "service": settings.SERVICE_NAME,
        "version": "2.0.0",
        "status": "running",
        "timestamp": datetime.now().isoformat(),
        "config": {
            "kafka_brokers": settings.KAFKA_BROKERS,
            "database_path": settings.DB_PATH,
            "log_level": settings.LOG_LEVEL,
            "debug": settings.DEBUG
        }
    }


if __name__ == "__main__":
    import uvicorn
    
    logger.info(f"Starting service on {settings.SERVICE_HOST}:{settings.SERVICE_PORT}")
    
    uvicorn.run(
        "main:app",
        host=settings.SERVICE_HOST,
        port=settings.SERVICE_PORT,
        reload=settings.DEBUG,
        log_level=settings.LOG_LEVEL.lower()
    )
