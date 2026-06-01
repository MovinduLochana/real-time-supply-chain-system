use axum::{
    http::StatusCode,
    response::{IntoResponse, Response},
    Json,
};
use serde_json::json;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum LogisticsError {
    #[error("Database error: {0}")]
    DatabaseError(String),

    #[error("Shipment not found")]
    ShipmentNotFound,

    #[error("Order not found")]
    #[allow(dead_code)]
    OrderNotFound,

    #[error("Location not found")]
    #[allow(dead_code)]
    LocationNotFound,

    #[error("Invalid input: {0}")]
    InvalidInput(String),

    #[error("Kafka error: {0}")]
    #[allow(dead_code)]
    KafkaError(String),

    #[error("Serialization error: {0}")]
    #[allow(dead_code)]
    SerializationError(String),

    #[error("Internal server error")]
    #[allow(dead_code)]
    InternalError,

    #[error("Conflict: {0}")]
    #[allow(dead_code)]
    Conflict(String),

    #[error("Failed to create shipment")]
    CreationFailed(String),
}

impl IntoResponse for LogisticsError {
    fn into_response(self) -> Response {
        let (status, error_message) = match &self {
            LogisticsError::DatabaseError(msg) => {
                tracing::error!("Database error: {}", msg);
                (
                    StatusCode::INTERNAL_SERVER_ERROR,
                    "Database error occurred".to_string(),
                )
            }
            LogisticsError::ShipmentNotFound => (StatusCode::NOT_FOUND, "Shipment not found".to_string()),
            LogisticsError::OrderNotFound => (StatusCode::NOT_FOUND, "Order not found".to_string()),
            LogisticsError::LocationNotFound => (StatusCode::NOT_FOUND, "Location not found".to_string()),
            LogisticsError::InvalidInput(msg) => {
                tracing::warn!("Invalid input: {}", msg);
                (StatusCode::BAD_REQUEST, msg.clone())
            }
            LogisticsError::KafkaError(msg) => {
                tracing::error!("Kafka error: {}", msg);
                (
                    StatusCode::INTERNAL_SERVER_ERROR,
                    "Message queue error occurred".to_string(),
                )
            }
            LogisticsError::SerializationError(msg) => {
                tracing::error!("Serialization error: {}", msg);
                (
                    StatusCode::INTERNAL_SERVER_ERROR,
                    "Serialization error occurred".to_string(),
                )
            }
            LogisticsError::InternalError => {
                tracing::error!("Internal server error");
                (StatusCode::INTERNAL_SERVER_ERROR, "Internal server error".to_string())
            }
            LogisticsError::Conflict(msg) => {
                tracing::warn!("Conflict: {}", msg);
                (StatusCode::CONFLICT, msg.clone())
            }
            LogisticsError::CreationFailed(msg) => {
                tracing::error!("Creation failed: {}", msg);
                (StatusCode::BAD_REQUEST, msg.clone())
            }
        };

        let body = Json(json!({
            "error": error_message,
            "status": status.as_u16()
        }));

        (status, body).into_response()
    }
}

pub type Result<T> = std::result::Result<T, LogisticsError>;
