package movindu.orderservice.dto;

import java.util.UUID;

public record OrderItem(
        UUID itemId,
        int quantity
) {
}
