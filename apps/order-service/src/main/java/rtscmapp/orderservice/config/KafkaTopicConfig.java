package rtscmapp.orderservice.config;

import org.apache.kafka.clients.admin.NewTopic;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

//@Configuration
public class KafkaTopicConfig {

//    @Bean
//    public NewTopic createOrderTopic(
//            @Value("${kafka.topic.order.name}") String topicName,
//            @Value("${kafka.topic.order.partitions}") int partitions,
//            @Value("${kafka.topic.order.replication-factor}") short replicationFactor) {
//        return new NewTopic(topicName, partitions, replicationFactor);
//    }

}
