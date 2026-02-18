package movindu.orderservice.entity;

import jakarta.persistence.*;
import lombok.Data;
import org.hibernate.annotations.CreationTimestamp;

import java.time.Instant;
import java.util.UUID;

@Entity
@Data
public class Item {
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    // Test cyclical dependency issue
    // If issues arise, remove this bidirectional association
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "order_id")
    private Order order;

    private String itemName;

    private int quantity;

    @Column(precision = 10, scale = 2)
    private double price;

    @CreationTimestamp
    private Instant createdAt;
}
