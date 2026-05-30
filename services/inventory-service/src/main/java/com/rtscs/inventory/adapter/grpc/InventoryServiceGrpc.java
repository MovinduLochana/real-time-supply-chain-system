package com.rtscs.inventory.adapter.grpc;

import com.rtscs.inventory.application.service.InventoryService;
import com.rtscs.proto.inventory.v1.*;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.server.service.GrpcService;

import java.util.ArrayList;
import java.util.List;

/**
 * gRPC service implementation for Inventory Service
 */
@GrpcService
@RequiredArgsConstructor
@Slf4j
public class InventoryServiceGrpc extends InventoryServiceGrpc.InventoryServiceImplBase {
    
    private final InventoryService inventoryService;
    
    @Override
    public void getStock(GetStockRequest request, StreamObserver<StockLevelData> responseObserver) {
        try {
            log.debug("gRPC: GetStock called for SKU: {}", request.getSku());
            
            String warehouseLocation = request.getWarehouseLocation().isEmpty() ? "DEFAULT" : request.getWarehouseLocation();
            StockLevelData response = inventoryService.getStock(request.getSku(), warehouseLocation);
            
            responseObserver.onNext(response);
            responseObserver.onCompleted();
            
        } catch (Exception e) {
            log.error("Error in GetStock: {}", e.getMessage(), e);
            responseObserver.onError(
                Status.INTERNAL
                    .withDescription(e.getMessage())
                    .asException()
            );
        }
    }
    
    @Override
    public void updateStock(UpdateStockRequest request, StreamObserver<UpdateStockResponse> responseObserver) {
        try {
            log.debug("gRPC: UpdateStock called for SKU: {} with delta: {}", request.getSku(), request.getQuantityDelta());
            
            String warehouseLocation = request.getWarehouseLocation().isEmpty() ? "DEFAULT" : request.getWarehouseLocation();
            StockLevelData updatedLevel = inventoryService.updateStock(
                request.getSku(),
                warehouseLocation,
                request.getQuantityDelta(),
                request.getReason()
            );
            
            UpdateStockResponse response = UpdateStockResponse.newBuilder()
                .setSuccess(true)
                .setUpdatedLevel(updatedLevel)
                .build();
            
            responseObserver.onNext(response);
            responseObserver.onCompleted();
            
        } catch (Exception e) {
            log.error("Error in UpdateStock: {}", e.getMessage(), e);
            responseObserver.onError(
                Status.INTERNAL
                    .withDescription(e.getMessage())
                    .asException()
            );
        }
    }
    
    @Override
    public void reserveStock(ReserveStockRequest request, StreamObserver<ReservationData> responseObserver) {
        try {
            log.debug("gRPC: ReserveStock called for order: {} with {} items", 
                request.getOrderId(), request.getItemsList().size());
            
            List<ReserveItem> items = new ArrayList<>(request.getItemsList());
            ReservationData response = inventoryService.reserveStock(request.getOrderId(), items);
            
            responseObserver.onNext(response);
            responseObserver.onCompleted();
            
        } catch (Exception e) {
            log.error("Error in ReserveStock: {}", e.getMessage(), e);
            responseObserver.onError(
                Status.INTERNAL
                    .withDescription(e.getMessage())
                    .asException()
            );
        }
    }
    
    @Override
    public void releaseReservation(ReleaseReservationRequest request, StreamObserver<ReleaseReservationResponse> responseObserver) {
        try {
            log.debug("gRPC: ReleaseReservation called for ID: {}", request.getReservationId());
            
            inventoryService.releaseReservation(request.getReservationId());
            
            ReleaseReservationResponse response = ReleaseReservationResponse.newBuilder()
                .setSuccess(true)
                .setMessage("Reservation released successfully")
                .build();
            
            responseObserver.onNext(response);
            responseObserver.onCompleted();
            
        } catch (Exception e) {
            log.error("Error in ReleaseReservation: {}", e.getMessage(), e);
            responseObserver.onError(
                Status.INTERNAL
                    .withDescription(e.getMessage())
                    .asException()
            );
        }
    }
    
    @Override
    public void getReservation(GetReservationRequest request, StreamObserver<ReservationData> responseObserver) {
        try {
            log.debug("gRPC: GetReservation called for ID: {}", request.getReservationId());
            responseObserver.onError(Status.UNIMPLEMENTED.asException());
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL.asException());
        }
    }
    
    @Override
    public void listItems(ListItemsRequest request, StreamObserver<ListItemsResponse> responseObserver) {
        try {
            log.debug("gRPC: ListItems called");
            responseObserver.onError(Status.UNIMPLEMENTED.asException());
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL.asException());
        }
    }
}
