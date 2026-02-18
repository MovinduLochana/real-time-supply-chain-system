package movindu.orderservice.controller;

import lombok.RequiredArgsConstructor;
import movindu.orderservice.dto.CreateOrderRequest;
import movindu.orderservice.dto.OrderResponse;
import movindu.orderservice.entity.Order;
import movindu.orderservice.entity.OrderStatus;
import movindu.orderservice.repo.OrderRepo;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.List;
import java.util.UUID;

@RestController
@RequestMapping("/api/v1/orders")
@RequiredArgsConstructor
public class OrderController {

    private final OrderRepo orderRepo;

    @GetMapping("/")
    public ResponseEntity<List<Order>> getAllOrders() {
        return ResponseEntity.ok(orderRepo.findAll());
    }

    @PostMapping("/create")
    public ResponseEntity<OrderResponse> createOrder(@AuthenticationPrincipal Jwt auth, CreateOrderRequest orderReq) {
        var customerKeycloakId = auth.getSubject();

        if(auth.getSubject() == null) return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();

        var order = Order.builder()
                .customerKeycloakId(UUID.fromString(customerKeycloakId))
                .items(orderReq.items())
                .orderStatus(OrderStatus.PENDING)
                .actualDelivery(Instant.now().plus(10, ChronoUnit.DAYS))
                .shippingAddress(orderReq.shippingAddress())
                .build();

        return OrderResponse.getOrderResponse(orderRepo.save(order));
    }
}
