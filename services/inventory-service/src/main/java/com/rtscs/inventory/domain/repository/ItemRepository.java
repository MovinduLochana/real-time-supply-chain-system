package com.rtscs.inventory.domain.repository;

import com.rtscs.inventory.domain.entity.Item;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface ItemRepository extends JpaRepository<Item, Long> {
    
    Optional<Item> findBySku(String sku);
    
    Page<Item> findByCategory(String category, Pageable pageable);
}
