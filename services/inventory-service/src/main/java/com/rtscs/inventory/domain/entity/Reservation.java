package com.rtscs.inventory.domain.entity;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;

/**
 * Reservation entity - tracks reserved inventory for orders
 */
@Entity
@Table(name = "reservations", indexes = {
    @Index(name = "idx_order_id", columnList = "order_id"),
    @Index(name = "idx_status", columnList = "status"),
    @Index(name = "idx_expires_at", columnList = "expires_at")
})
@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class Reservation {
    
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private String id;
    
    @Column(nullable = false, length = 100)
    private String orderId;
    
    @ManyToOne(fetch = FetchType.LAZY, optional = false)
    @JoinColumn(name = "item_id", nullable = false)
    private Item item;
    
    @Column(nullable = false)
    private Integer quantityReserved;
    
    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private ReservationStatus status;
    
    @Column(nullable = false)
    private Instant createdAt;
    
    @Column(nullable = false)
    private Instant expiresAt;
    
    @Column
    private Instant releasedAt;
    
    @PrePersist
    protected void onCreate() {
        createdAt = Instant.now();
        if (expiresAt == null) {
            // Default expiry: 24 hours
            expiresAt = Instant.now().plusSeconds(86400);
        }
    }
    
    public enum ReservationStatus {
        PENDING,
        CONFIRMED,
        RELEASED,
        EXPIRED
    }
    
    /**
     * Check if reservation has expired
     */
    public boolean isExpired() {
        return Instant.now().isAfter(expiresAt) && status != ReservationStatus.RELEASED;
    }
}
