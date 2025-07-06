package rtscmapp.orderservice.model;

import jakarta.persistence.*;
import lombok.Data;
import org.hibernate.annotations.CreationTimestamp;
import rtscmapp.orderservice.enums.OrderStatus;

import java.time.LocalDateTime;

@Data @Entity
public class IOrder {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private int id;

    @CreationTimestamp
    private LocalDateTime createdAt;

    private OrderStatus status = OrderStatus.PENDING;

//    List<OrderItem> items;

}
