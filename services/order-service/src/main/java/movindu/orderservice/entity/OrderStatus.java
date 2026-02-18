package movindu.orderservice.entity;

public enum OrderStatus {
    PENDING,
    CONFIRMED,
    PROCESSING,
    SHIPPED,
    IN_TRANSIT,
    OUT_FOR_DELIVERY,
    CANCELLED,
    RETURNED,
    REFUNDED
}
