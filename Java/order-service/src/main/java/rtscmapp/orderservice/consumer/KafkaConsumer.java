package rtscmapp.orderservice.consumer;

import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Component;
import rtscmapp.orderservice.model.IOrder;

@Component
public class KafkaConsumer {

//    @KafkaListener(topics = "order-topic", groupId = "order-group")
//    public void consume(String message) {
//        System.out.println("Received message: " + message);
//    }

    @KafkaListener(topics = "order-topic", groupId = "order-group")
    public void consume(IOrder order) {
        System.out.println("Received message: " + order);
    }

}
