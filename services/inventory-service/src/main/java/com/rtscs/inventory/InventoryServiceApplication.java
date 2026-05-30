package com.rtscs.inventory;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cache.annotation.EnableCaching;
import org.springframework.scheduling.annotation.EnableScheduling;

/**
 * Inventory Service - Main Spring Boot Application
 * 
 * Responsibilities:
 * - Manage product catalog and stock levels
 * - Handle inventory reservations for orders
 * - Publish inventory events to Kafka
 * - Expose gRPC API for other services
 */
@SpringBootApplication
@EnableCaching
@EnableScheduling
public class InventoryServiceApplication {

    public static void main(String[] args) {
        SpringApplication.run(InventoryServiceApplication.class, args);
    }
}
