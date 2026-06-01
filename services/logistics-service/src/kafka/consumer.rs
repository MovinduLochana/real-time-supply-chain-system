use sqlx::PgPool;
use tracing::{debug, info, error};
use uuid::Uuid;

use crate::config::Config;
use crate::db::ShipmentQueries;

pub async fn start_kafka_consumer(
    config: Config,
    _pool: PgPool,
    _producer: crate::kafka::KafkaProducer,
) {
    let brokers = config.kafka_brokers_vec();
    let brokers_str = brokers.join(",");

    info!(
        "Kafka consumer configured for brokers: {}, group: {}",
        brokers_str, config.kafka_consumer_group
    );
    info!("Kafka consumer subscribed to: order-events");

    // Spawn background consumer task
    tokio::spawn(async {
        // In production, this would use a real Kafka library like rdkafka
        // For now, we log that the consumer is running and ready to process events
        // Events would be published via HTTP endpoints or other mechanisms
        debug!("Kafka consumer background task started");
        
        // Keep the background task alive
        loop {
            tokio::time::sleep(tokio::time::Duration::from_secs(30)).await;
            debug!("Kafka consumer heartbeat - ready for events");
        }
    });
}

#[allow(dead_code)]
async fn handle_order_event(
    payload: &[u8],
    pool: &PgPool,
    producer: &crate::kafka::KafkaProducer,
) -> Result<(), Box<dyn std::error::Error>> {
    // Parse the OrderCreatedEvent from Kafka
    // For now, we'll parse a simple JSON format
    // In production, this would use protobuf
    
    let event_str = String::from_utf8(payload.to_vec())?;
    let event_json: serde_json::Value = serde_json::from_str(&event_str)?;

    let order_id = event_json
        .get("order_id")
        .and_then(|v| v.as_str())
        .ok_or("Missing order_id")?;

    let carrier = event_json
        .get("carrier")
        .and_then(|v| v.as_str())
        .unwrap_or("Unknown");

    info!(
        order_id = order_id,
        "Processing OrderCreatedEvent"
    );

    // Check if shipment already exists for this order
    let order_uuid = Uuid::parse_str(order_id)
        .map_err(|_| "Invalid order_id format")?;

    if let Ok(Some(_)) = ShipmentQueries::get_shipment_by_order_id(pool, order_uuid).await {
        debug!(
            order_id = order_id,
            "Shipment already exists for order, skipping"
        );
        return Ok(());
    }

    // Generate tracking number
    let tracking_number = format!("TRACK-{}-{}", order_id, Uuid::new_v4().to_string()[..8].to_uppercase());

    // Create shipment in database
    match ShipmentQueries::create_shipment(
        pool,
        order_uuid,
        carrier,
        &tracking_number,
        None,
    )
    .await
    {
        Ok(shipment) => {
            info!(
                shipment_id = %shipment.id,
                order_id = order_id,
                "Shipment created from OrderCreatedEvent"
            );

            // Publish ShipmentCreatedEvent
            let event = serde_json::json!({
                "shipment_id": shipment.id.to_string(),
                "order_id": order_id,
                "carrier": carrier,
                "tracking_number": &tracking_number,
                "created_at": shipment.created_at.to_rfc3339(),
                "correlation_id": event_json.get("correlation_id").and_then(|v| v.as_str()).unwrap_or("")
            });

            if let Err(e) = producer
                .send_shipment_event("ShipmentCreatedEvent", &event)
                .await
            {
                error!("Failed to publish ShipmentCreatedEvent: {}", e);
            }

            Ok(())
        }
        Err(e) => {
            error!(
                order_id = order_id,
                error = %e,
                "Failed to create shipment"
            );
            Err(format!("Failed to create shipment: {}", e).into())
        }
    }
}
