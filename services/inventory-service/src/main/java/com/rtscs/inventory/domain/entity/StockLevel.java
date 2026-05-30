package com.rtscs.inventory.domain.entity;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;

/**
 * StockLevel entity - tracks inventory quantity at warehouses
 */
@Entity
@Table(name = "stock_levels", indexes = {
    @Index(name = "idx_item_location", columnList = "item_id,warehouse_location", unique = true),
    @Index(name = "idx_warehouse", columnList = "warehouse_location")
})
@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class StockLevel {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @ManyToOne(fetch = FetchType.LAZY, optional = false)
    @JoinColumn(name = "item_id", nullable = false)
    private Item item;
    
    @Column(nullable = false, length = 50)
    private String warehouseLocation;
    
    @Column(nullable = false)
    private Integer quantityOnHand;
    
    @Column(nullable = false)
    private Integer quantityReserved;
    
    @Column(nullable = false)
    private Instant lastCountedAt;
    
    @Column(nullable = false, updatable = false)
    private Instant createdAt;
    
    @Column(nullable = false)
    private Instant updatedAt;
    
    @PrePersist
    protected void onCreate() {
        createdAt = Instant.now();
        updatedAt = Instant.now();
    }
    
    @PreUpdate
    protected void onUpdate() {
        updatedAt = Instant.now();
    }
    
    /**
     * Calculate available quantity (on-hand minus reserved)
     */
    public Integer getQuantityAvailable() {
        return quantityOnHand - quantityReserved;
    }
    
    /**
     * Check if sufficient stock is available
     */
    public boolean hasAvailableQuantity(Integer requestedQuantity) {
        return getQuantityAvailable() >= requestedQuantity;
    }
}
