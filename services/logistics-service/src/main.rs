mod config;
mod db;
mod error;
mod handlers;
mod kafka;
mod routes;
mod telemetry;

use axum::middleware::Next;
use axum::response::Response;
use axum::{
    extract::Request,
    middleware,
};
use sqlx::postgres::PgPoolOptions;
use std::net::SocketAddr;
use tower::ServiceBuilder;
use tower_http::cors::CorsLayer;
use tower_http::trace::{DefaultMakeSpan, DefaultOnResponse, TraceLayer};
use tracing::Level;

use config::Config;
use handlers::ShipmentState;
use kafka::{start_kafka_consumer, KafkaProducer};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Load configuration
    let config = Config::from_env();

    // Initialize tracing
    telemetry::init_tracing_simple();

    tracing::info!(
        service = %config.service_name,
        port = config.service_port,
        "Starting Logistics Service"
    );

    // Initialize database connection pool
    let pool = PgPoolOptions::new()
        .max_connections(20)
        .connect(&config.database_url)
        .await?;

    tracing::info!("Database connection pool created");

    // Run migrations
    sqlx::migrate!("./migrations")
        .run(&pool)
        .await
        .map_err(|e| {
            tracing::error!("Migration failed: {}", e);
            e
        })?;

    tracing::info!("Database migrations completed");

    // Initialize Kafka producer
    let kafka_brokers = config.kafka_brokers_vec();
    let kafka_producer = KafkaProducer::new(&kafka_brokers)
        .map_err(|e| {
            tracing::error!("Failed to create Kafka producer: {}", e);
            e
        })?;

    // Start Kafka consumer in background
    let kafka_consumer_config = config.clone();
    let kafka_consumer_pool = pool.clone();
    let kafka_consumer_producer = kafka_producer.clone();

    start_kafka_consumer(kafka_consumer_config, kafka_consumer_pool, kafka_consumer_producer).await;

    tracing::info!("Kafka consumer started");

    // Create application state
    let state = ShipmentState {
        db: pool,
        kafka: kafka_producer,
    };

    // Create router
    let app = routes::create_router(state)
        .layer(
            ServiceBuilder::new()
                .layer(
                    TraceLayer::new_for_http()
                        .make_span_with(DefaultMakeSpan::new().level(Level::INFO))
                        .on_response(DefaultOnResponse::new().level(Level::INFO)),
                )
                .layer(middleware::from_fn(add_request_id))
                .layer(CorsLayer::permissive())
        );

    // Start server
    let addr = SocketAddr::from(([0, 0, 0, 0], config.service_port));
    let listener = tokio::net::TcpListener::bind(&addr).await?;

    tracing::info!(
        addr = %addr,
        "Server listening"
    );

    axum::serve(listener, app).await?;

    Ok(())
}

async fn add_request_id(mut request: Request, next: Next) -> Response {
    let request_id = uuid::Uuid::new_v4().to_string();
    request.extensions_mut().insert(request_id.clone());

    tracing::debug!(
        request_id = %request_id,
        method = %request.method(),
        uri = %request.uri(),
        "Incoming request"
    );

    next.run(request).await
}
