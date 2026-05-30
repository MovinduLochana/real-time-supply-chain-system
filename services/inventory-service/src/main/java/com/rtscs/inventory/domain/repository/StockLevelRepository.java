package com.rtscs.inventory.domain.repository;

import com.rtscs.inventory.domain.entity.StockLevel;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Lock;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import jakarta.persistence.LockModeType;
import java.util.Optional;

@Repository
public interface StockLevelRepository extends JpaRepository<StockLevel, Long> {
    
    @Lock(LockModeType.PESSIMISTIC_WRITE)
    @Query("SELECT s FROM StockLevel s WHERE s.item.sku = :sku AND s.warehouseLocation = :warehouseLocation")
    Optional<StockLevel> findBySkuAndWarehouseLocationWithLock(
        @Param("sku") String sku,
        @Param("warehouseLocation") String warehouseLocation
    );
    
    @Query("SELECT s FROM StockLevel s WHERE s.item.sku = :sku AND s.warehouseLocation = :warehouseLocation")
    Optional<StockLevel> findBySkuAndWarehouseLocation(
        @Param("sku") String sku,
        @Param("warehouseLocation") String warehouseLocation
    );
}
