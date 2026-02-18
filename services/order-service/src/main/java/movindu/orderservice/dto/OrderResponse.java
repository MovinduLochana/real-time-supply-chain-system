package movindu.orderservice.dto;

import movindu.orderservice.entity.Order;
import movindu.orderservice.entity.OrderStatus;
import org.jspecify.annotations.NonNull;
import org.springframework.http.ResponseEntity;

import java.util.UUID;

public record OrderResponse(
        UUID orderId,
        OrderStatus status
) {
    public static ResponseEntity<@NonNull OrderResponse> getOrderResponse(Order order) {
        return ResponseEntity.ok(new OrderResponse(order.getOrderId(), order.getOrderStatus()));
    }
}
