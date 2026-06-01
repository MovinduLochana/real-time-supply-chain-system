use chrono::Utc;
use sqlx::PgPool;
use uuid::Uuid;

use crate::error::{LogisticsError, Result};

use super::models::{Shipment, ShipmentLocation};

pub struct ShipmentQueries;

impl ShipmentQueries {
    pub async fn create_shipment(
        pool: &PgPool,
        order_id: Uuid,
        carrier: &str,
        tracking_number: &str,
        destination_address: Option<&str>,
    ) -> Result<Shipment> {
        let shipment_id = Uuid::new_v4();
        let now = Utc::now();

        let shipment = sqlx::query_as::<_, Shipment>(
            r#"
            INSERT INTO shipments (id, order_id, carrier, tracking_number, status, destination_address, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            RETURNING *
            "#
        )
        .bind(shipment_id)
        .bind(order_id)
        .bind(carrier)
        .bind(tracking_number)
        .bind("CREATED")
        .bind(destination_address)
        .bind(now)
        .bind(now)
        .fetch_one(pool)
        .await
        .map_err(|e| {
            tracing::error!("Failed to create shipment: {}", e);
            LogisticsError::CreationFailed(format!("Database error: {}", e))
        })?;

        tracing::info!(
            shipment_id = %shipment.id,
            order_id = %shipment.order_id,
            "Shipment created successfully"
        );

        Ok(shipment)
    }

    pub async fn get_shipment(pool: &PgPool, shipment_id: Uuid) -> Result<Shipment> {
        let shipment = sqlx::query_as::<_, Shipment>(
            "SELECT * FROM shipments WHERE id = $1"
        )
        .bind(shipment_id)
        .fetch_optional(pool)
        .await
        .map_err(|e| {
            tracing::error!("Failed to get shipment: {}", e);
            LogisticsError::DatabaseError(e.to_string())
        })?
        .ok_or(LogisticsError::ShipmentNotFound)?;

        Ok(shipment)
    }

    #[allow(dead_code)]
    pub async fn get_shipment_by_order_id(
        pool: &PgPool,
        order_id: Uuid,
    ) -> Result<Option<Shipment>> {
        let shipment = sqlx::query_as::<_, Shipment>(
            "SELECT * FROM shipments WHERE order_id = $1 LIMIT 1"
        )
        .bind(order_id)
        .fetch_optional(pool)
        .await
        .map_err(|e| {
            tracing::error!("Failed to get shipment by order_id: {}", e);
            LogisticsError::DatabaseError(e.to_string())
        })?;

        Ok(shipment)
    }

    pub async fn update_shipment_status(
        pool: &PgPool,
        shipment_id: Uuid,
        new_status: &str,
    ) -> Result<Shipment> {
        let now = Utc::now();

        let shipment = sqlx::query_as::<_, Shipment>(
            r#"
            UPDATE shipments
            SET status = $1, updated_at = $2
            WHERE id = $3
            RETURNING *
            "#
        )
        .bind(new_status)
        .bind(now)
        .bind(shipment_id)
        .fetch_optional(pool)
        .await
        .map_err(|e| {
            tracing::error!("Failed to update shipment status: {}", e);
            LogisticsError::DatabaseError(e.to_string())
        })?
        .ok_or(LogisticsError::ShipmentNotFound)?;

        tracing::info!(
            shipment_id = %shipment.id,
            new_status = %new_status,
            "Shipment status updated"
        );

        Ok(shipment)
    }

    pub async fn update_shipment_location(
        pool: &PgPool,
        shipment_id: Uuid,
        latitude: f64,
        longitude: f64,
        accuracy_meters: Option<f64>,
        address: Option<&str>,
    ) -> Result<Shipment> {
        let now = Utc::now();

        let shipment = sqlx::query_as::<_, Shipment>(
            r#"
            UPDATE shipments
            SET current_latitude = $1, current_longitude = $2, current_accuracy_meters = $3, current_address = $4, updated_at = $5
            WHERE id = $6
            RETURNING *
            "#
        )
        .bind(latitude)
        .bind(longitude)
        .bind(accuracy_meters)
        .bind(address)
        .bind(now)
        .bind(shipment_id)
        .fetch_optional(pool)
        .await
        .map_err(|e| {
            tracing::error!("Failed to update shipment location: {}", e);
            LogisticsError::DatabaseError(e.to_string())
        })?
        .ok_or(LogisticsError::ShipmentNotFound)?;

        tracing::debug!(
            shipment_id = %shipment.id,
            latitude = latitude,
            longitude = longitude,
            "Shipment location updated"
        );

        Ok(shipment)
    }
}

pub struct LocationQueries;

impl LocationQueries {
    pub async fn create_location(
        pool: &PgPool,
        shipment_id: Uuid,
        latitude: f64,
        longitude: f64,
        accuracy_meters: Option<f64>,
        address: Option<&str>,
    ) -> Result<ShipmentLocation> {
        let location_id = Uuid::new_v4();
        let now = Utc::now();

        let location = sqlx::query_as::<_, ShipmentLocation>(
            r#"
            INSERT INTO shipment_locations (id, shipment_id, latitude, longitude, accuracy_meters, address, recorded_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7)
            RETURNING *
            "#
        )
        .bind(location_id)
        .bind(shipment_id)
        .bind(latitude)
        .bind(longitude)
        .bind(accuracy_meters)
        .bind(address)
        .bind(now)
        .fetch_one(pool)
        .await
        .map_err(|e| {
            tracing::error!("Failed to create location: {}", e);
            LogisticsError::CreationFailed(format!("Database error: {}", e))
        })?;

        Ok(location)
    }

    pub async fn get_locations(
        pool: &PgPool,
        shipment_id: Uuid,
        limit: i64,
    ) -> Result<Vec<ShipmentLocation>> {
        let locations = sqlx::query_as::<_, ShipmentLocation>(
            r#"
            SELECT * FROM shipment_locations
            WHERE shipment_id = $1
            ORDER BY recorded_at DESC
            LIMIT $2
            "#
        )
        .bind(shipment_id)
        .bind(limit)
        .fetch_all(pool)
        .await
        .map_err(|e| {
            tracing::error!("Failed to get locations: {}", e);
            LogisticsError::DatabaseError(e.to_string())
        })?;

        Ok(locations)
    }

    #[allow(dead_code)]
    pub async fn get_latest_location(
        pool: &PgPool,
        shipment_id: Uuid,
    ) -> Result<Option<ShipmentLocation>> {
        let location = sqlx::query_as::<_, ShipmentLocation>(
            r#"
            SELECT * FROM shipment_locations
            WHERE shipment_id = $1
            ORDER BY recorded_at DESC
            LIMIT 1
            "#
        )
        .bind(shipment_id)
        .fetch_optional(pool)
        .await
        .map_err(|e| {
            tracing::error!("Failed to get latest location: {}", e);
            LogisticsError::DatabaseError(e.to_string())
        })?;

        Ok(location)
    }
}
