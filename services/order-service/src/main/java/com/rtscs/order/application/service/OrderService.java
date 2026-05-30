package com.rtscs.order.application.service;

import com.rtscs.order.domain.entity.OrderEntity;
import com.rtscs.order.domain.entity.OrderLineItem;
import com.rtscs.order.domain.repository.OrderRepository;
import com.rtscs.proto.inventory.v1.ReserveItem;
import com.rtscs.proto.inventory.v1.ReserveStockRequest;
import com.rtscs.proto.inventory.v1.InventoryServiceGrpc;
import com.rtscs.proto.order.v1.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

/**
 * Order Service - Core business logic
 * Orchestrates order creation with inventory reservation via gRPC
 */
@Service
@RequiredArgsConstructor
@Slf4j
@Transactional
public class OrderService {
    
    private final OrderRepository orderRepository;
    private final InventoryServiceGrpc.InventoryServiceBlockingStub inventoryServiceStub;
    private final KafkaTemplate<String, Object> kafkaTemplate;
    
    private static final String TOPIC_ORDER_EVENTS = "order-events";
    
    /**
     * Create an order and reserve inventory
     */
    public Order createOrder(CreateOrderRequest request) {
        log.info("Creating order for customer: {}", request.getCustomerId());
        
        try {
            // Prepare inventory reservation request
            List<ReserveItem> reserveItems = new ArrayList<>();
            double totalAmount = 0;
            
            for (CreateOrderRequest.OrderLineItem item : request.getLineItemsList()) {
                reserveItems.add(ReserveItem.newBuilder()
                    .setSku(item.getSku())
                    .setQuantity(item.getQuantity())
                    .build());
                totalAmount += item.getUnitPrice() * item.getQuantity();
            }
            
            // Call Inventory Service via gRPC to reserve stock
            log.debug("Calling InventoryService.ReserveStock for {} items", reserveItems.size());
            
            ReserveStockRequest reserveRequest = ReserveStockRequest.newBuilder()
                .setOrderId(UUID.randomUUID().toString())
                .addAllItems(reserveItems)
                .build();
            
            var reservationResponse = inventoryServiceStub.reserveStock(reserveRequest);
            log.info("Stock reservation successful: {}", reservationResponse.getReservationId());
            
            // Create order entity
            OrderEntity order = OrderEntity.builder()
                .id(UUID.randomUUID().toString())
                .customerId(request.getCustomerId())
                .status(OrderEntity.OrderStatus.PENDING)
                .totalAmount(totalAmount)
                .shippingAddress(request.getShippingAddress())
                .build();
            
            // Create line items
            List<OrderLineItem> lineItems = new ArrayList<>();
            for (CreateOrderRequest.OrderLineItem item : request.getLineItemsList()) {
                OrderLineItem lineItem = OrderLineItem.builder()
                    .order(order)
                    .itemSku(item.getSku())
                    .quantity(item.getQuantity())
                    .unitPrice(item.getUnitPrice())
                    .build();
                lineItems.add(lineItem);
            }
            order.setLineItems(lineItems);
            
            // Save order
            order = orderRepository.save(order);
            log.info("Order created: {}", order.getId());
            
            // Publish OrderCreated event to Kafka
            publishOrderCreatedEvent(order);
            
            return mapOrderToProto(order);
            
        } catch (Exception e) {
            log.error("Error creating order: {}", e.getMessage(), e);
            throw new OrderException("Failed to create order: " + e.getMessage(), e);
        }
    }
    
    /**
     * Get order by ID
     */
    @Transactional(readOnly = true)
    public Order getOrder(String orderId) {
        log.debug("Getting order: {}", orderId);
        
        OrderEntity order = orderRepository.findById(orderId)
            .orElseThrow(() -> new OrderException("Order not found: " + orderId));
        
        return mapOrderToProto(order);
    }
    
    /**
     * Cancel order and release reservations
     */
    public Order cancelOrder(String orderId, String reason) {
        log.info("Cancelling order: {} - Reason: {}", orderId, reason);
        
        OrderEntity order = orderRepository.findById(orderId)
            .orElseThrow(() -> new OrderException("Order not found: " + orderId));
        
        order.setStatus(OrderEntity.OrderStatus.CANCELLED);
        order = orderRepository.save(order);
        
        // Publish event
        publishOrderCancelledEvent(order, reason);
        
        return mapOrderToProto(order);
    }
    
    /**
     * Map Order entity to protobuf
     */
    private Order mapOrderToProto(OrderEntity order) {
        Order.Builder builder = Order.newBuilder()
            .setOrderId(order.getId())
            .setCustomerId(order.getCustomerId())
            .setTotalAmount(order.getTotalAmount())
            .setStatus(mapOrderStatus(order.getStatus()))
            .setShippingAddress(order.getShippingAddress())
            .setCreatedAt(com.google.protobuf.Timestamp.newBuilder()
                .setSeconds(order.getCreatedAt().getEpochSecond())
                .build());
        
        for (OrderLineItem item : order.getLineItems()) {
            builder.addLineItems(LineItem.newBuilder()
                .setItemSku(item.getItemSku())
                .setItemName(item.getItemSku())
                .setQuantity(item.getQuantity())
                .setUnitPrice(item.getUnitPrice())
                .setSubtotal(item.getSubtotal())
                .build());
        }
        
        return builder.build();
    }
    
    private Order.Status mapOrderStatus(OrderEntity.OrderStatus status) {
        return switch (status) {
            case PENDING -> Order.Status.PENDING;
            case CONFIRMED -> Order.Status.CONFIRMED;
            case SHIPPED -> Order.Status.SHIPPED;
            case DELIVERED -> Order.Status.DELIVERED;
            case CANCELLED -> Order.Status.CANCELLED;
            case FAILED -> Order.Status.FAILED;
        };
    }
    
    private void publishOrderCreatedEvent(OrderEntity order) {
        OrderEvent event = OrderEvent.newBuilder()
            .setOrderId(order.getId())
            .setEventType(OrderEvent.EventType.ORDER_CREATED)
            .setOrder(mapOrderToProto(order))
            .setEventTimestamp(com.google.protobuf.Timestamp.newBuilder()
                .setSeconds(Instant.now().getEpochSecond())
                .build())
            .setCorrelationId(UUID.randomUUID().toString())
            .build();
        
        kafkaTemplate.send(TOPIC_ORDER_EVENTS, order.getId(), event);
        log.info("Published OrderCreatedEvent for order: {}", order.getId());
    }
    
    private void publishOrderCancelledEvent(OrderEntity order, String reason) {
        OrderEvent event = OrderEvent.newBuilder()
            .setOrderId(order.getId())
            .setEventType(OrderEvent.EventType.ORDER_CANCELLED)
            .setOrder(mapOrderToProto(order))
            .setEventTimestamp(com.google.protobuf.Timestamp.newBuilder()
                .setSeconds(Instant.now().getEpochSecond())
                .build())
            .setCorrelationId(UUID.randomUUID().toString())
            .build();
        
        kafkaTemplate.send(TOPIC_ORDER_EVENTS, order.getId(), event);
        log.info("Published OrderCancelledEvent for order: {}", order.getId());
    }
}

/**
 * Order-specific exception
 */
class OrderException extends RuntimeException {
    public OrderException(String message) {
        super(message);
    }
    
    public OrderException(String message, Throwable cause) {
        super(message, cause);
    }
}
