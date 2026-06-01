use axum::{
    extract::{Path, State},
    http::StatusCode,
    Json,
};
use chrono::Utc;
use uuid::Uuid;

use crate::db::{models::*, ShipmentQueries};
use crate::error::Result;
use crate::kafka::KafkaProducer;
use sqlx::PgPool;
use tracing::info;

#[derive(Clone)]
pub struct ShipmentState {
    pub db: PgPool,
    pub kafka: KafkaProducer,
}

pub async fn create_shipment(
    State(state): State<ShipmentState>,
    Json(payload): Json<CreateShipmentRequest>,
) -> Result<(StatusCode, Json<ShipmentResponse>)> {
    let order_id = Uuid::parse_str(&payload.order_id)
        .map_err(|_| crate::error::LogisticsError::InvalidInput("Invalid order_id format".to_string()))?;

    let tracking_number = format!(
        "TRACK-{}-{}",
        payload.order_id,
        uuid::Uuid::new_v4().to_string()[..8].to_uppercase()
    );

    let shipment = ShipmentQueries::create_shipment(
        &state.db,
        order_id,
        &payload.carrier,
        &tracking_number,
        payload.destination_address.as_deref(),
    )
    .await?;

    info!(
        shipment_id = %shipment.id,
        order_id = %shipment.order_id,
        "Shipment created via API"
    );

    // Publish ShipmentCreatedEvent to Kafka
    let event = serde_json::json!({
        "shipment_id": shipment.id.to_string(),
        "order_id": shipment.order_id.to_string(),
        "carrier": shipment.carrier,
        "tracking_number": shipment.tracking_number,
        "created_at": shipment.created_at.to_rfc3339(),
        "correlation_id": uuid::Uuid::new_v4().to_string()
    });

    if let Err(e) = state.kafka.send_shipment_event("ShipmentCreatedEvent", &event).await {
        tracing::warn!("Failed to publish ShipmentCreatedEvent: {}", e);
    }

    Ok((StatusCode::CREATED, Json(shipment.into())))
}

pub async fn get_shipment(
    State(state): State<ShipmentState>,
    Path(shipment_id): Path<String>,
) -> Result<Json<ShipmentResponse>> {
    let shipment_uuid = Uuid::parse_str(&shipment_id)
        .map_err(|_| crate::error::LogisticsError::InvalidInput("Invalid shipment_id format".to_string()))?;

    let shipment = ShipmentQueries::get_shipment(&state.db, shipment_uuid).await?;

    Ok(Json(shipment.into()))
}

pub async fn update_shipment_status(
    State(state): State<ShipmentState>,
    Path(shipment_id): Path<String>,
    Json(payload): Json<UpdateShipmentStatusRequest>,
) -> Result<Json<StatusUpdateResponse>> {
    let shipment_uuid = Uuid::parse_str(&shipment_id)
        .map_err(|_| crate::error::LogisticsError::InvalidInput("Invalid shipment_id format".to_string()))?;

    // Get the current shipment to get previous status
    let current_shipment = ShipmentQueries::get_shipment(&state.db, shipment_uuid).await?;
    let previous_status = current_shipment.status.clone();

    // Validate status
    if ShipmentStatus::from_string(&payload.status).is_none() {
        return Err(crate::error::LogisticsError::InvalidInput(format!(
            "Invalid status: {}",
            payload.status
        )));
    }

    // Update the status
    let updated_shipment =
        ShipmentQueries::update_shipment_status(&state.db, shipment_uuid, &payload.status).await?;

    info!(
        shipment_id = %shipment_uuid,
        previous_status = %previous_status,
        new_status = %payload.status,
        "Shipment status updated"
    );

    // Publish ShipmentStatusChangedEvent
    let event = serde_json::json!({
        "shipment_id": shipment_uuid.to_string(),
        "order_id": updated_shipment.order_id.to_string(),
        "previous_status": previous_status,
        "new_status": payload.status,
        "reason": payload.reason.unwrap_or_default(),
        "changed_at": Utc::now().to_rfc3339(),
        "correlation_id": uuid::Uuid::new_v4().to_string()
    });

    if let Err(e) = state
        .kafka
        .send_shipment_event("ShipmentStatusChangedEvent", &event)
        .await
    {
        tracing::warn!("Failed to publish ShipmentStatusChangedEvent: {}", e);
    }

    Ok(Json(StatusUpdateResponse {
        shipment_id: shipment_uuid.to_string(),
        previous_status,
        new_status: payload.status,
        updated_at: updated_shipment.updated_at,
    }))
}
