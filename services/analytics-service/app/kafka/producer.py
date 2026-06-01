"""Kafka producer for analytics events."""
import logging
import json
from typing import Dict, Any, Optional
from aio_pika import connect_robust
from datetime import datetime
from uuid import uuid4
from config import settings


logger = logging.getLogger(__name__)


class KafkaProducer:
    """Kafka producer for publishing analytics events."""
    
    def __init__(self):
        """Initialize Kafka producer."""
        self.brokers = settings.KAFKA_BROKERS
        self.connection = None
        self.channel = None
    
    async def connect(self) -> None:
        """Connect to Kafka."""
        try:
            logger.info(f"Connecting to Kafka: {self.brokers}")
            self.connection = await connect_robust(f"amqp://guest:guest@{self.brokers.split(':')[0]}/")
            self.channel = await self.connection.channel()
            logger.info("Producer connected to Kafka")
        except Exception as e:
            logger.error(f"Failed to connect producer: {e}")
            raise
    
    async def disconnect(self) -> None:
        """Disconnect from Kafka."""
        if self.connection:
            await self.connection.close()
            logger.info("Producer disconnected from Kafka")
    
    async def publish_metrics_event(self, metrics: Dict[str, Any]) -> bool:
        """Publish metrics aggregated event."""
        try:
            event = {
                "event_id": str(uuid4()),
                "event_type": "MetricsAggregatedEvent",
                "timestamp": datetime.now().isoformat(),
                "period": metrics.get("period", "hourly"),
                "metrics": metrics
            }
            
            return await self._publish_event(
                settings.KAFKA_METRICS_TOPIC,
                event
            )
        except Exception as e:
            logger.error(f"Failed to publish metrics event: {e}")
            return False
    
    async def publish_forecast_event(self, forecast: Dict[str, Any]) -> bool:
        """Publish forecast generated event."""
        try:
            event = {
                "event_id": str(uuid4()),
                "event_type": "ForecastGeneratedEvent",
                "sku": forecast.get("sku"),
                "days": forecast.get("forecast_days"),
                "forecast_data": forecast.get("forecast", []),
                "model_type": forecast.get("model_type", "prophet"),
                "timestamp": datetime.now().isoformat()
            }
            
            return await self._publish_event(
                settings.KAFKA_FORECAST_TOPIC,
                event
            )
        except Exception as e:
            logger.error(f"Failed to publish forecast event: {e}")
            return False
    
    async def _publish_event(self, topic: str, event: Dict[str, Any]) -> bool:
        """Publish event to topic."""
        try:
            if not self.channel:
                await self.connect()
            
            exchange = await self.channel.declare_exchange(
                name=topic,
                auto_delete=True
            )
            
            message_body = json.dumps(event).encode()
            from aio_pika import Message
            message = Message(message_body)
            
            await exchange.publish(message, routing_key="")
            
            logger.debug(f"Published event to {topic}: {event.get('event_id')}")
            return True
        
        except Exception as e:
            logger.error(f"Failed to publish to {topic}: {e}")
            return False


# Global instance
_producer: Optional[KafkaProducer] = None


def get_kafka_producer() -> KafkaProducer:
    """Get Kafka producer instance."""
    global _producer
    if _producer is None:
        _producer = KafkaProducer()
    return _producer
