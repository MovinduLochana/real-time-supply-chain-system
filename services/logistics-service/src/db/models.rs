use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, sqlx::Type)]
#[sqlx(type_name = "varchar")]
pub enum ShipmentStatus {
    #[serde(rename = "CREATED")]
    Created,
    #[serde(rename = "IN_TRANSIT")]
    InTransit,
    #[serde(rename = "OUT_FOR_DELIVERY")]
    OutForDelivery,
    #[serde(rename = "DELIVERED")]
    Delivered,
    #[serde(rename = "FAILED")]
    Failed,
    #[serde(rename = "CANCELLED")]
    Cancelled,
}

impl ShipmentStatus {
    #[allow(dead_code)]
    pub fn to_string(&self) -> String {
        match self {
            ShipmentStatus::Created => "CREATED".to_string(),
            ShipmentStatus::InTransit => "IN_TRANSIT".to_string(),
            ShipmentStatus::OutForDelivery => "OUT_FOR_DELIVERY".to_string(),
            ShipmentStatus::Delivered => "DELIVERED".to_string(),
            ShipmentStatus::Failed => "FAILED".to_string(),
            ShipmentStatus::Cancelled => "CANCELLED".to_string(),
        }
    }

    pub fn from_string(s: &str) -> Option<ShipmentStatus> {
        match s {
            "CREATED" => Some(ShipmentStatus::Created),
            "IN_TRANSIT" => Some(ShipmentStatus::InTransit),
            "OUT_FOR_DELIVERY" => Some(ShipmentStatus::OutForDelivery),
            "DELIVERED" => Some(ShipmentStatus::Delivered),
            "FAILED" => Some(ShipmentStatus::Failed),
            "CANCELLED" => Some(ShipmentStatus::Cancelled),
            _ => None,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct Shipment {
    pub id: Uuid,
    pub order_id: Uuid,
    pub carrier: String,
    pub tracking_number: String,
    pub status: String,
    pub destination_address: Option<String>,
    pub current_latitude: Option<f64>,
    pub current_longitude: Option<f64>,
    pub current_accuracy_meters: Option<f64>,
    pub current_address: Option<String>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ShipmentResponse {
    pub shipment_id: String,
    pub order_id: String,
    pub carrier: String,
    pub tracking_number: String,
    pub status: String,
    pub destination_address: Option<String>,
    pub current_location: Option<LocationResponse>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

impl From<Shipment> for ShipmentResponse {
    fn from(s: Shipment) -> Self {
        let current_location = if let (Some(lat), Some(lng)) = (s.current_latitude, s.current_longitude) {
            Some(LocationResponse {
                latitude: lat,
                longitude: lng,
                accuracy_meters: s.current_accuracy_meters,
                address: s.current_address,
                recorded_at: s.updated_at,
            })
        } else {
            None
        };

        ShipmentResponse {
            shipment_id: s.id.to_string(),
            order_id: s.order_id.to_string(),
            carrier: s.carrier,
            tracking_number: s.tracking_number,
            status: s.status,
            destination_address: s.destination_address,
            current_location,
            created_at: s.created_at,
            updated_at: s.updated_at,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct ShipmentLocation {
    pub id: Uuid,
    pub shipment_id: Uuid,
    pub latitude: f64,
    pub longitude: f64,
    pub accuracy_meters: Option<f64>,
    pub address: Option<String>,
    pub recorded_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LocationResponse {
    pub latitude: f64,
    pub longitude: f64,
    pub accuracy_meters: Option<f64>,
    pub address: Option<String>,
    pub recorded_at: DateTime<Utc>,
}

impl From<ShipmentLocation> for LocationResponse {
    fn from(l: ShipmentLocation) -> Self {
        LocationResponse {
            latitude: l.latitude,
            longitude: l.longitude,
            accuracy_meters: l.accuracy_meters,
            address: l.address,
            recorded_at: l.recorded_at,
        }
    }
}

#[derive(Debug, Serialize, Deserialize)]
pub struct CreateShipmentRequest {
    pub order_id: String,
    pub carrier: String,
    pub destination_address: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct UpdateLocationRequest {
    pub latitude: f64,
    pub longitude: f64,
    pub accuracy_meters: Option<f64>,
    pub address: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct UpdateShipmentStatusRequest {
    pub status: String,
    pub reason: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct LocationHistoryResponse {
    pub locations: Vec<LocationResponse>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct CreateLocationResponse {
    pub success: bool,
    pub shipment_id: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct StatusUpdateResponse {
    pub shipment_id: String,
    pub previous_status: String,
    pub new_status: String,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct HealthResponse {
    pub status: String,
    pub service: String,
    pub timestamp: DateTime<Utc>,
}
