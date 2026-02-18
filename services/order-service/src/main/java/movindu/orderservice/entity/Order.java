package movindu.orderservice.entity;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.Immutable;
import org.hibernate.annotations.JdbcTypeCode;
import org.hibernate.type.SqlTypes;
import org.jspecify.annotations.Nullable;
import org.springframework.data.annotation.LastModifiedDate;

import java.time.Instant;
import java.util.List;
import java.util.Map;
import java.util.UUID;

@Entity
@Immutable
@Data
@AllArgsConstructor
@NoArgsConstructor
public class Order {
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID orderId;

    @Column(nullable = false, updatable = false)
    private UUID customerKeycloakId;

    @Enumerated(EnumType.STRING)
    private OrderStatus orderStatus;

    private double orderAmount;

    @JdbcTypeCode(SqlTypes.JSON)
    private Map<String, String> shippingAddress;

    // remove if bidirectional association removed from item table
    @OneToMany(mappedBy = "order", fetch = FetchType.EAGER)
    private List<Item> items;

    @CreationTimestamp
    private Instant createdAt;

    @LastModifiedDate
    private Instant updatedAt;

    private Instant estimatedDelivery;

    @Nullable
    private Instant actualDelivery;

    @PrePersist
    public void prePersist() {
        this.orderAmount = !items.isEmpty() ? items.stream().mapToDouble(Item::getPrice).sum() : 0;
    }

    @Builder
    public Order(UUID customerKeycloakId, OrderStatus orderStatus, Map<String, String> shippingAddress, List<Item> items, Instant estimatedDelivery, @Nullable Instant actualDelivery) {
        this.customerKeycloakId = customerKeycloakId;
        this.orderStatus = orderStatus;
        this.shippingAddress = shippingAddress;
        this.items = items;
        this.estimatedDelivery = estimatedDelivery;
        this.actualDelivery = actualDelivery;
    }
}
