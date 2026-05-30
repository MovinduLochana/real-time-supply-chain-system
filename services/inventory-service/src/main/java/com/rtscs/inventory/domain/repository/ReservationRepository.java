package com.rtscs.inventory.domain.repository;

import com.rtscs.inventory.domain.entity.Reservation;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.Instant;
import java.util.List;
import java.util.Optional;

@Repository
public interface ReservationRepository extends JpaRepository<Reservation, String> {
    
    Optional<Reservation> findByOrderId(String orderId);
    
    List<Reservation> findByOrderId(String orderId);
    
    @Query("SELECT r FROM Reservation r WHERE r.expiresAt < :now AND r.status = 'PENDING'")
    List<Reservation> findExpiredReservations(@Param("now") Instant now);
}
