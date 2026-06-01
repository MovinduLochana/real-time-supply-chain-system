use axum::{
    extract::State,
    http::StatusCode,
    response::IntoResponse,
    routing::{get, patch, post},
    Json, Router,
};
use chrono::Utc;

use crate::db::models::HealthResponse;
use crate::handlers::{
    create_location, create_shipment, get_locations, get_shipment, update_shipment_status,
    ShipmentState,
};

pub fn create_router(state: ShipmentState) -> Router {
    Router::new()
        .route("/shipments", post(create_shipment))
        .route("/shipments/:shipment_id", get(get_shipment))
        .route(
            "/shipments/:shipment_id/locations",
            post(create_location).get(get_locations),
        )
        .route(
            "/shipments/:shipment_id/status",
            patch(update_shipment_status),
        )
        .route("/health", get(health_check))
        .with_state(state)
}

async fn health_check(State(_state): State<ShipmentState>) -> impl IntoResponse {
    (
        StatusCode::OK,
        Json(HealthResponse {
            status: "healthy".to_string(),
            service: "logistics-service".to_string(),
            timestamp: Utc::now(),
        }),
    )
}
