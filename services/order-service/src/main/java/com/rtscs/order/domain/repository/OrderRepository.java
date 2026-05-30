package com.rtscs.order.domain.repository;

import com.rtscs.order.domain.entity.OrderEntity;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface OrderRepository extends JpaRepository<OrderEntity, String> {
    
    Optional<OrderEntity> findById(String id);
    
    Page<OrderEntity> findByCustomerId(String customerId, Pageable pageable);
    
    Page<OrderEntity> findByStatus(OrderEntity.OrderStatus status, Pageable pageable);
}
