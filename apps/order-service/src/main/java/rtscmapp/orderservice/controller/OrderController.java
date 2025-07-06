package rtscmapp.orderservice.controller;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import rtscmapp.orderservice.model.IOrder;
import rtscmapp.orderservice.repo.IOrderRepo;

import java.util.List;

@RestController
@RequestMapping("/api/v1/orders")
public class OrderController {

    private final IOrderRepo orderRepo;

    public OrderController(IOrderRepo orderRepo) {
        this.orderRepo = orderRepo;
    }

    @GetMapping("/")
    public ResponseEntity<List<IOrder>> getAllOrders() {
        return ResponseEntity.ok().body(orderRepo.findAll());
    }

    @GetMapping("/{id}")
    public ResponseEntity<IOrder> getOrder(@PathVariable int id) {
        return ResponseEntity.ok().body(orderRepo.findById(id).orElseThrow());
    }

    @PostMapping("/")
    public ResponseEntity<IOrder> getAllOrders(@RequestBody IOrder order) {
        return ResponseEntity.ok().body(orderRepo.save(order));
    }

}
