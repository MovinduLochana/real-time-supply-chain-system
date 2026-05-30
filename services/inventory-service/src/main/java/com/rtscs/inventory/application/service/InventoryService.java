package com.rtscs.inventory.application.service;

import com.rtscs.inventory.domain.entity.Item;
import com.rtscs.inventory.domain.entity.Reservation;
import com.rtscs.inventory.domain.entity.StockLevel;
import com.rtscs.inventory.domain.repository.ItemRepository;
import com.rtscs.inventory.domain.repository.ReservationRepository;
import com.rtscs.inventory.domain.repository.StockLevelRepository;
import com.rtscs.proto.inventory.v1.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.cache.annotation.Cacheable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.*;

/**
 * Inventory Service - Core business logic
 */
@Service
@RequiredArgsConstructor
@Slf4j
@Transactional
public class InventoryService {
    
    private final ItemRepository itemRepository;
    private final StockLevelRepository stockLevelRepository;
    private final ReservationRepository reservationRepository;
    private final InventoryEventPublisher eventPublisher;
    
    /**
     * Get stock level for an item
     */
    @Transactional(readOnly = true)
    @Cacheable(value = "stockLevels", key = "#sku + ':' + #warehouseLocation", unless = "#result == null")
    public StockLevelData getStock(String sku, String warehouseLocation) {
        log.debug("Getting stock for SKU: {} at location: {}", sku, warehouseLocation);
        
        StockLevel stockLevel = stockLevelRepository
            .findBySkuAndWarehouseLocation(sku, warehouseLocation)
            .orElseThrow(() -> new InventoryException("Stock level not found for SKU: " + sku));
        
        return mapStockLevelToData(stockLevel);
    }
    
    /**
     * Update stock quantity
     */
    public StockLevelData updateStock(String sku, String warehouseLocation, Integer quantityDelta, String reason) {
        log.info("Updating stock for SKU: {} at location: {}, delta: {}, reason: {}", 
            sku, warehouseLocation, quantityDelta, reason);
        
        StockLevel stockLevel = stockLevelRepository
            .findBySkuAndWarehouseLocationWithLock(sku, warehouseLocation)
            .orElseThrow(() -> new InventoryException("Stock level not found for SKU: " + sku));
        
        int previousQuantity = stockLevel.getQuantityOnHand();
        int newQuantity = previousQuantity + quantityDelta;
        
        if (newQuantity < 0) {
            throw new InventoryException("Insufficient stock. Current: " + previousQuantity + ", Delta: " + quantityDelta);
        }
        
        stockLevel.setQuantityOnHand(newQuantity);
        stockLevel.setLastCountedAt(Instant.now());
        stockLevel = stockLevelRepository.save(stockLevel);
        
        // Publish event
        eventPublisher.publishStockUpdatedEvent(sku, warehouseLocation, previousQuantity, newQuantity, reason);
        
        log.info("Stock updated for SKU: {} from {} to {}", sku, previousQuantity, newQuantity);
        return mapStockLevelToData(stockLevel);
    }
    
    /**
     * Reserve stock for an order
     */
    public ReservationData reserveStock(String orderId, List<ReserveItem> items) {
        log.info("Reserving stock for order: {}", orderId);
        
        // Check if reservation already exists
        Optional<Reservation> existingReservation = reservationRepository.findByOrderId(orderId);
        if (existingReservation.isPresent()) {
            throw new InventoryException("Reservation already exists for order: " + orderId);
        }
        
        // Reserve each item
        List<Reservation> reservations = new ArrayList<>();
        try {
            for (ReserveItem item : items) {
                StockLevel stockLevel = stockLevelRepository
                    .findBySkuAndWarehouseLocationWithLock(item.getSku(), "DEFAULT")
                    .orElseThrow(() -> new InventoryException("Item not found: " + item.getSku()));
                
                if (!stockLevel.hasAvailableQuantity(item.getQuantity())) {
                    throw new InventoryException(
                        String.format("Insufficient stock for %s. Available: %d, Requested: %d",
                            item.getSku(),
                            stockLevel.getQuantityAvailable(),
                            item.getQuantity())
                    );
                }
                
                // Update reserved quantity
                stockLevel.setQuantityReserved(stockLevel.getQuantityReserved() + item.getQuantity());
                stockLevelRepository.save(stockLevel);
                
                // Create reservation
                Reservation reservation = Reservation.builder()
                    .orderId(orderId)
                    .item(stockLevel.getItem())
                    .quantityReserved(item.getQuantity())
                    .status(Reservation.ReservationStatus.PENDING)
                    .build();
                
                reservation = reservationRepository.save(reservation);
                reservations.add(reservation);
                
                log.info("Reserved {} units of SKU: {} for order: {}", 
                    item.getQuantity(), item.getSku(), orderId);
                
                // Publish event
                eventPublisher.publishStockReservedEvent(orderId, item.getSku(), item.getQuantity());
            }
        } catch (Exception e) {
            // Release all previously reserved stock on error
            releaseAllReservations(orderId);
            throw e;
        }
        
        // Return first reservation as response (in real scenario, might return all or aggregate)
        return mapReservationToData(reservations.get(0));
    }
    
    /**
     * Release a reservation
     */
    public void releaseReservation(String reservationId) {
        log.info("Releasing reservation: {}", reservationId);
        
        Reservation reservation = reservationRepository.findById(reservationId)
            .orElseThrow(() -> new InventoryException("Reservation not found: " + reservationId));
        
        if (reservation.getStatus() == Reservation.ReservationStatus.RELEASED) {
            throw new InventoryException("Reservation already released: " + reservationId);
        }
        
        // Update stock level
        Item item = reservation.getItem();
        StockLevel stockLevel = stockLevelRepository
            .findBySkuAndWarehouseLocationWithLock(item.getSku(), "DEFAULT")
            .orElseThrow(() -> new InventoryException("Stock level not found"));
        
        stockLevel.setQuantityReserved(stockLevel.getQuantityReserved() - reservation.getQuantityReserved());
        stockLevelRepository.save(stockLevel);
        
        // Update reservation
        reservation.setStatus(Reservation.ReservationStatus.RELEASED);
        reservation.setReleasedAt(Instant.now());
        reservationRepository.save(reservation);
        
        // Publish event
        eventPublisher.publishStockReleasedEvent(reservation.getOrderId(), item.getSku(), reservation.getQuantityReserved());
        
        log.info("Reservation released: {}", reservationId);
    }
    
    /**
     * Release all reservations for an order
     */
    private void releaseAllReservations(String orderId) {
        List<Reservation> reservations = reservationRepository.findByOrderId(orderId);
        for (Reservation reservation : reservations) {
            if (reservation.getStatus() != Reservation.ReservationStatus.RELEASED) {
                try {
                    releaseReservation(reservation.getId());
                } catch (Exception e) {
                    log.error("Failed to release reservation: {}", reservation.getId(), e);
                }
            }
        }
    }
    
    /**
     * Map entity to protobuf data
     */
    private StockLevelData mapStockLevelToData(StockLevel stockLevel) {
        return StockLevelData.newBuilder()
            .setItemSku(stockLevel.getItem().getSku())
            .setWarehouseLocation(stockLevel.getWarehouseLocation())
            .setQuantityOnHand(stockLevel.getQuantityOnHand())
            .setQuantityReserved(stockLevel.getQuantityReserved())
            .setQuantityAvailable(stockLevel.getQuantityAvailable())
            .setLastCountedAt(com.google.protobuf.Timestamp.newBuilder()
                .setSeconds(stockLevel.getLastCountedAt().getEpochSecond())
                .setNanos(stockLevel.getLastCountedAt().getNano())
                .build())
            .build();
    }
    
    private ReservationData mapReservationToData(Reservation reservation) {
        return ReservationData.newBuilder()
            .setReservationId(reservation.getId())
            .setOrderId(reservation.getOrderId())
            .setItemSku(reservation.getItem().getSku())
            .setQuantityReserved(reservation.getQuantityReserved())
            .setStatus(mapReservationStatus(reservation.getStatus()))
            .setCreatedAt(com.google.protobuf.Timestamp.newBuilder()
                .setSeconds(reservation.getCreatedAt().getEpochSecond())
                .setNanos(reservation.getCreatedAt().getNano())
                .build())
            .setExpiresAt(com.google.protobuf.Timestamp.newBuilder()
                .setSeconds(reservation.getExpiresAt().getEpochSecond())
                .setNanos(reservation.getExpiresAt().getNano())
                .build())
            .build();
    }
    
    private ReservationData.Status mapReservationStatus(Reservation.ReservationStatus status) {
        return switch (status) {
            case PENDING -> ReservationData.Status.PENDING;
            case CONFIRMED -> ReservationData.Status.CONFIRMED;
            case RELEASED -> ReservationData.Status.RELEASED;
            case EXPIRED -> ReservationData.Status.EXPIRED;
        };
    }
}

/**
 * Custom exception for inventory errors
 */
class InventoryException extends RuntimeException {
    public InventoryException(String message) {
        super(message);
    }
    
    public InventoryException(String message, Throwable cause) {
        super(message, cause);
    }
}
