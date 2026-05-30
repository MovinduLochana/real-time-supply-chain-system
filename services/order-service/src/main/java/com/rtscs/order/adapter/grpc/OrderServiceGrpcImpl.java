package com.rtscs.order.adapter.grpc;

import com.rtscs.order.application.service.OrderService;
import com.rtscs.proto.order.v1.*;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.server.service.GrpcService;

/**
 * gRPC service implementation for Order Service
 */
@GrpcService
@RequiredArgsConstructor
@Slf4j
public class OrderServiceGrpcImpl extends OrderServiceGrpc.OrderServiceImplBase {
    
    private final OrderService orderService;
    
    @Override
    public void createOrder(CreateOrderRequest request, StreamObserver<Order> responseObserver) {
        try {
            log.debug("gRPC: CreateOrder called for customer: {}", request.getCustomerId());
            
            Order response = orderService.createOrder(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
            
        } catch (Exception e) {
            log.error("Error in CreateOrder: {}", e.getMessage(), e);
            responseObserver.onError(
                Status.INTERNAL
                    .withDescription(e.getMessage())
                    .asException()
            );
        }
    }
    
    @Override
    public void getOrder(GetOrderRequest request, StreamObserver<Order> responseObserver) {
        try {
            log.debug("gRPC: GetOrder called for ID: {}", request.getOrderId());
            
            Order response = orderService.getOrder(request.getOrderId());
            responseObserver.onNext(response);
            responseObserver.onCompleted();
            
        } catch (Exception e) {
            log.error("Error in GetOrder: {}", e.getMessage(), e);
            responseObserver.onError(
                Status.NOT_FOUND
                    .withDescription(e.getMessage())
                    .asException()
            );
        }
    }
    
    @Override
    public void listOrders(ListOrdersRequest request, StreamObserver<ListOrdersResponse> responseObserver) {
        try {
            log.debug("gRPC: ListOrders called");
            responseObserver.onError(Status.UNIMPLEMENTED.asException());
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL.asException());
        }
    }
    
    @Override
    public void cancelOrder(CancelOrderRequest request, StreamObserver<CancelOrderResponse> responseObserver) {
        try {
            log.debug("gRPC: CancelOrder called for ID: {}", request.getOrderId());
            
            Order cancelledOrder = orderService.cancelOrder(request.getOrderId(), request.getReason());
            
            CancelOrderResponse response = CancelOrderResponse.newBuilder()
                .setSuccess(true)
                .setCancelledOrder(cancelledOrder)
                .build();
            
            responseObserver.onNext(response);
            responseObserver.onCompleted();
            
        } catch (Exception e) {
            log.error("Error in CancelOrder: {}", e.getMessage(), e);
            responseObserver.onError(
                Status.INTERNAL
                    .withDescription(e.getMessage())
                    .asException()
            );
        }
    }
}
