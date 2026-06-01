"""Kafka consumer for analytics events."""
import logging
import json
import asyncio
from typing import Callable, Optional, Dict, Any
from aio_pika import connect_robust, IncomingMessage
from datetime import datetime
from config import settings
from app.models.schemas import (
    OrderCreatedEvent,
    OrderCancelledEvent,
    StockUpdatedEvent,
    LowStockAlertEvent
)


logger = logging.getLogger(__name__)


class KafkaConsumer:
    """Kafka consumer for event consumption."""
    
    def __init__(self):
        """Initialize Kafka consumer."""
        self.brokers = settings.KAFKA_BROKERS
        self.connection = None
        self.channel = None
        self.exchange = None
        self.event_handlers: Dict[str, Callable] = {}
        self._running = False
    
    async def connect(self) -> None:
        """Connect to Kafka broker."""
        try:
            logger.info(f"Connecting to Kafka: {self.brokers}")
            self.connection = await connect_robust(f"amqp://guest:guest@{self.brokers.split(':')[0]}/")
            self.channel = await self.connection.channel()
            logger.info("Connected to Kafka")
        except Exception as e:
            logger.error(f"Failed to connect to Kafka: {e}")
            raise
    
    async def disconnect(self) -> None:
        """Disconnect from Kafka."""
        self._running = False
        if self.connection:
            await self.connection.close()
            logger.info("Disconnected from Kafka")
    
    def register_handler(self, event_type: str, handler: Callable) -> None:
        """Register event handler."""
        self.event_handlers[event_type] = handler
        logger.info(f"Registered handler for event type: {event_type}")
    
    async def consume_order_events(self) -> None:
        """Consume order events."""
        try:
            if not self.channel:
                await self.connect()
            
            # Declare queue
            queue = await self.channel.declare_queue(
                name=f"{settings.KAFKA_CONSUMER_GROUP}_orders",
                durable=True
            )
            
            # Bind to exchange
            exchange = await self.channel.declare_exchange(
                name=settings.KAFKA_ORDER_EVENTS_TOPIC,
                auto_delete=True
            )
            await queue.bind(exchange)
            
            logger.info(f"Consuming from {settings.KAFKA_ORDER_EVENTS_TOPIC}")
            
            self._running = True
            async with queue.iterator() as queue_iter:
                async for message: IncomingMessage in queue_iter:
                    if not self._running:
                        break
                    
                    try:
                        async with message.process():
                            await self._process_order_event(message)
                    except Exception as e:
                        logger.error(f"Failed to process order event: {e}")
        
        except Exception as e:
            logger.error(f"Order consumer error: {e}")
            raise
    
    async def consume_inventory_events(self) -> None:
        """Consume inventory events."""
        try:
            if not self.channel:
                await self.connect()
            
            # Declare queue
            queue = await self.channel.declare_queue(
                name=f"{settings.KAFKA_CONSUMER_GROUP}_inventory",
                durable=True
            )
            
            # Bind to exchange
            exchange = await self.channel.declare_exchange(
                name=settings.KAFKA_INVENTORY_EVENTS_TOPIC,
                auto_delete=True
            )
            await queue.bind(exchange)
            
            logger.info(f"Consuming from {settings.KAFKA_INVENTORY_EVENTS_TOPIC}")
            
            self._running = True
            async with queue.iterator() as queue_iter:
                async for message: IncomingMessage in queue_iter:
                    if not self._running:
                        break
                    
                    try:
                        async with message.process():
                            await self._process_inventory_event(message)
                    except Exception as e:
                        logger.error(f"Failed to process inventory event: {e}")
        
        except Exception as e:
            logger.error(f"Inventory consumer error: {e}")
            raise
    
    async def _process_order_event(self, message: IncomingMessage) -> None:
        """Process order event."""
        try:
            body = json.loads(message.body.decode())
            event_type = body.get("event_type")
            
            if event_type == "OrderCreatedEvent":
                event = OrderCreatedEvent(**body)
                handler = self.event_handlers.get("OrderCreatedEvent")
                if handler:
                    await handler(event)
            
            elif event_type == "OrderCancelledEvent":
                event = OrderCancelledEvent(**body)
                handler = self.event_handlers.get("OrderCancelledEvent")
                if handler:
                    await handler(event)
            
            logger.debug(f"Processed order event: {event_type}")
        
        except Exception as e:
            logger.error(f"Error processing order event: {e}")
    
    async def _process_inventory_event(self, message: IncomingMessage) -> None:
        """Process inventory event."""
        try:
            body = json.loads(message.body.decode())
            event_type = body.get("event_type")
            
            if event_type == "StockUpdatedEvent":
                event = StockUpdatedEvent(**body)
                handler = self.event_handlers.get("StockUpdatedEvent")
                if handler:
                    await handler(event)
            
            elif event_type == "LowStockAlertEvent":
                event = LowStockAlertEvent(**body)
                handler = self.event_handlers.get("LowStockAlertEvent")
                if handler:
                    await handler(event)
            
            logger.debug(f"Processed inventory event: {event_type}")
        
        except Exception as e:
            logger.error(f"Error processing inventory event: {e}")


# Global instance
_consumer: Optional[KafkaConsumer] = None


def get_kafka_consumer() -> KafkaConsumer:
    """Get Kafka consumer instance."""
    global _consumer
    if _consumer is None:
        _consumer = KafkaConsumer()
    return _consumer
