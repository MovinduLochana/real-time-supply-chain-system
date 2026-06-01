use std::sync::Arc;
use tokio::sync::Mutex;
use tracing::{debug, info};

/// KafkaProducer provides a cross-platform Kafka event publishing interface
/// For production use with Windows/Linux, this can be replaced with rdkafka or kafka-rs
#[derive(Clone)]
pub struct KafkaProducer {
    #[allow(dead_code)]
    brokers: Vec<String>,
    events: Arc<Mutex<Vec<KafkaEvent>>>,
}

#[derive(Clone, Debug)]
#[allow(dead_code)]
struct KafkaEvent {
    topic: String,
    key: String,
    payload: String,
    event_type: String,
}

impl KafkaProducer {
    pub fn new(brokers: &[&str]) -> Result<Self, String> {
        let brokers_str = brokers.iter().map(|s| s.to_string()).collect::<Vec<_>>();
        info!("Kafka producer created, brokers: {}", brokers_str.join(","));

        Ok(KafkaProducer {
            brokers: brokers_str,
            events: Arc::new(Mutex::new(Vec::new())),
        })
    }

    pub async fn send_shipment_event(
        &self,
        event_type: &str,
        event: &serde_json::Value,
    ) -> Result<(), String> {
        let key = format!(
            "shipment-{}",
            event
                .get("shipment_id")
                .and_then(|v| v.as_str())
                .unwrap_or("unknown")
        );

        let payload = serde_json::to_string(event)
            .map_err(|e| format!("Failed to serialize event: {}", e))?;

        let kafka_event = KafkaEvent {
            topic: "shipment-events".to_string(),
            key: key.clone(),
            payload: payload.clone(),
            event_type: event_type.to_string(),
        };

        // In production, this would send to real Kafka
        // For now, we log the event and store it in the queue
        self.events.lock().await.push(kafka_event);

        debug!(
            event_type = event_type,
            key = key,
            "Event published to Kafka"
        );

        Ok(())
    }

    pub async fn send_location_event(
        &self,
        event: &serde_json::Value,
    ) -> Result<(), String> {
        let key = format!(
            "location-{}",
            event
                .get("shipment_id")
                .and_then(|v| v.as_str())
                .unwrap_or("unknown")
        );

        let payload = serde_json::to_string(event)
            .map_err(|e| format!("Failed to serialize event: {}", e))?;

        let kafka_event = KafkaEvent {
            topic: "shipment-events".to_string(),
            key: key.clone(),
            payload: payload.clone(),
            event_type: "LocationUpdatedEvent".to_string(),
        };

        // In production, this would send to real Kafka
        self.events.lock().await.push(kafka_event);

        debug!(
            key = key,
            "Location event published to Kafka"
        );

        Ok(())
    }

    #[allow(dead_code)]
    pub async fn flush(&self) -> Result<(), String> {
        // In production, flush would ensure all messages are sent
        // For now, this is a no-op
        Ok(())
    }

    /// Get pending events (for testing/monitoring purposes)
    #[allow(dead_code)]
    pub async fn get_pending_events(&self) -> Vec<KafkaEvent> {
        self.events.lock().await.clone()
    }

    /// Clear pending events (for testing purposes)
    #[allow(dead_code)]
    pub async fn clear_events(&self) {
        self.events.lock().await.clear();
    }
}
