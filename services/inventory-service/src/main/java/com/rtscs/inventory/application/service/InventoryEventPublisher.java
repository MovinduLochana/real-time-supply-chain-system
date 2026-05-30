package com.rtscs.inventory.application.service;

import com.rtscs.proto.events.v1.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Component;

import java.time.Instant;
import java.util.UUID;

/**
 * Publishes inventory events to Kafka
 */
@Component
@RequiredArgsConstructor
@Slf4j
public class InventoryEventPublisher {
    
    private final KafkaTemplate<String, Object> kafkaTemplate;
    private static final String TOPIC_INVENTORY_EVENTS = "inventory-events";
    
    public void publishStockUpdatedEvent(String sku, String warehouse, int previousQty, int newQty, String reason) {
        StockUpdatedEvent event = StockUpdatedEvent.newBuilder()
            .setSku(sku)
            .setWarehouseLocation(warehouse)
            .setPreviousQuantity(previousQty)
            .setNewQuantity(newQty)
            .setQuantityDelta(newQty - previousQty)
            .setReason(reason)
            .setEventTimestamp(com.google.protobuf.Timestamp.newBuilder()
                .setSeconds(Instant.now().getEpochSecond())
                .setNanos(Instant.now().getNano())
                .build())
            .setCorrelationId(UUID.randomUUID().toString())
            .build();
        
        kafkaTemplate.send(TOPIC_INVENTORY_EVENTS, sku, event);
        log.info("Published StockUpdatedEvent for SKU: {} ({}  -> {})", sku, previousQty, newQty);
    }
    
    public void publishStockReservedEvent(String orderId, String sku, int quantity) {
        StockReservedEvent event = StockReservedEvent.newBuilder()
            .setReservationId(UUID.randomUUID().toString())
            .setOrderId(orderId)
            .setSku(sku)
            .setQuantityReserved(quantity)
            .setReservedAt(com.google.protobuf.Timestamp.newBuilder()
                .setSeconds(Instant.now().getEpochSecond())
                .setNanos(Instant.now().getNano())
                .build())
            .setCorrelationId(UUID.randomUUID().toString())
            .build();
        
        kafkaTemplate.send(TOPIC_INVENTORY_EVENTS, orderId, event);
        log.info("Published StockReservedEvent for order: {}, SKU: {}, qty: {}", orderId, sku, quantity);
    }
    
    public void publishStockReleasedEvent(String orderId, String sku, int quantity) {
        StockReleasedEvent event = StockReleasedEvent.newBuilder()
            .setReservationId(UUID.randomUUID().toString())
            .setOrderId(orderId)
            .setSku(sku)
            .setQuantityReleased(quantity)
            .setReleasedAt(com.google.protobuf.Timestamp.newBuilder()
                .setSeconds(Instant.now().getEpochSecond())
                .setNanos(Instant.now().getNano())
                .build())
            .setCorrelationId(UUID.randomUUID().toString())
            .build();
        
        kafkaTemplate.send(TOPIC_INVENTORY_EVENTS, orderId, event);
        log.info("Published StockReleasedEvent for order: {}, SKU: {}, qty: {}", orderId, sku, quantity);
    }
}
