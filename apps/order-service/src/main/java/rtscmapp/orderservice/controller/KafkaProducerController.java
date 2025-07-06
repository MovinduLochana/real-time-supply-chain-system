package rtscmapp.orderservice.controller;

import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.web.bind.annotation.*;
import rtscmapp.orderservice.model.IOrder;

@RestController
@RequestMapping("/api/v1/kafka")
public class KafkaProducerController {

    //private final KafkaTemplate<String, String> kafkaTemplate;
    private final KafkaTemplate<String, IOrder> kafkaOrderTemplate;

    public KafkaProducerController( KafkaTemplate<String, IOrder> kafkaOrderTemplate) {
        //this.kafkaTemplate = kafkaTemplate;
        this.kafkaOrderTemplate = kafkaOrderTemplate;
    }

//    @PostMapping("/send")
//    public String sendMessage(@RequestParam String message) {
//        kafkaTemplate.send("order-topic", message);
//        return message;
//    }

    @PostMapping("/send-order")
    public IOrder sendSerializedOrder(@RequestBody IOrder order) {
        System.out.println(order.getId());
        kafkaOrderTemplate.send("order-topic", order);
        return order;
    }

}
