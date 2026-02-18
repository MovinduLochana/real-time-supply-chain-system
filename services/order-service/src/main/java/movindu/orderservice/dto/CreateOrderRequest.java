package movindu.orderservice.dto;

import movindu.orderservice.entity.Item;

import java.util.List;
import java.util.Map;
import java.util.UUID;

public record CreateOrderRequest(
        UUID itemId,
        UUID customerKeycloakId,
        List<Item> items,
        Map<String, String> shippingAddress
) {

}
