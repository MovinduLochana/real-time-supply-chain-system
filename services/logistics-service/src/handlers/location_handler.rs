use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    Json,
};
use chrono::Utc;
use serde::Deserialize;
use uuid::Uuid;

use crate::db::{models::*, LocationQueries, ShipmentQueries};
use crate::error::Result;
use crate::handlers::ShipmentState;
use tracing::info;

#[derive(Deserialize)]
pub struct LocationQuery {
    limit: Option<i64>,
}

pub async fn create_location(
    State(state): State<ShipmentState>,
    Path(shipment_id): Path<String>,
    Json(payload): Json<UpdateLocationRequest>,
) -> Result<(StatusCode, Json<CreateLocationResponse>)> {
    let shipment_uuid = Uuid::parse_str(&shipment_id)
        .map_err(|_| crate::error::LogisticsError::InvalidInput("Invalid shipment_id format".to_string()))?;

    // Verify shipment exists
    let _shipment = ShipmentQueries::get_shipment(&state.db, shipment_uuid).await?;

    // Create location record
    LocationQueries::create_location(
        &state.db,
        shipment_uuid,
        payload.latitude,
        payload.longitude,
        payload.accuracy_meters,
        payload.address.as_deref(),
    )
    .await?;

    // Update shipment with latest location
    ShipmentQueries::update_shipment_location(
        &state.db,
        shipment_uuid,
        payload.latitude,
        payload.longitude,
        payload.accuracy_meters,
        payload.address.as_deref(),
    )
    .await?;

    info!(
        shipment_id = %shipment_uuid,
        latitude = payload.latitude,
        longitude = payload.longitude,
        "Location updated"
    );

    // Publish LocationUpdatedEvent
    let event = serde_json::json!({
        "shipment_id": shipment_uuid.to_string(),
        "latitude": payload.latitude,
        "longitude": payload.longitude,
        "accuracy_meters": payload.accuracy_meters.unwrap_or(0.0),
        "address": payload.address.unwrap_or_default(),
        "recorded_at": Utc::now().to_rfc3339(),
        "correlation_id": uuid::Uuid::new_v4().to_string()
    });

    if let Err(e) = state.kafka.send_location_event(&event).await {
        tracing::warn!("Failed to publish LocationUpdatedEvent: {}", e);
    }

    Ok((
        StatusCode::CREATED,
        Json(CreateLocationResponse {
            success: true,
            shipment_id: shipment_uuid.to_string(),
        }),
    ))
}

pub async fn get_locations(
    State(state): State<ShipmentState>,
    Path(shipment_id): Path<String>,
    Query(params): Query<LocationQuery>,
) -> Result<Json<LocationHistoryResponse>> {
    let shipment_uuid = Uuid::parse_str(&shipment_id)
        .map_err(|_| crate::error::LogisticsError::InvalidInput("Invalid shipment_id format".to_string()))?;

    // Verify shipment exists
    let _shipment = ShipmentQueries::get_shipment(&state.db, shipment_uuid).await?;

    let limit = params.limit.unwrap_or(100);
    if limit <= 0 || limit > 1000 {
        return Err(crate::error::LogisticsError::InvalidInput(
            "Limit must be between 1 and 1000".to_string(),
        ));
    }

    let locations = LocationQueries::get_locations(&state.db, shipment_uuid, limit).await?;

    let location_responses = locations
        .into_iter()
        .map(LocationResponse::from)
        .collect();

    Ok(Json(LocationHistoryResponse {
        locations: location_responses,
    }))
}
