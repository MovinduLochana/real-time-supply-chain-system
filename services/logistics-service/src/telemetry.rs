use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

#[allow(dead_code)]
pub fn init_tracing(service_name: &str) -> Result<(), Box<dyn std::error::Error>> {
    let jaeger_agent_host = std::env::var("JAEGER_AGENT_HOST").unwrap_or_else(|_| "localhost".to_string());
    let jaeger_agent_port: u16 = std::env::var("JAEGER_AGENT_PORT")
        .unwrap_or_else(|_| "6831".to_string())
        .parse()?;

    let tracer = opentelemetry_jaeger::new_agent_pipeline()
        .with_service_name(service_name)
        .with_auto_split_batch(true)
        .with_endpoint((jaeger_agent_host.as_str(), jaeger_agent_port))
        .install_simple()?;

    let telemetry = tracing_opentelemetry::layer().with_tracer(tracer);

    let env_filter = tracing_subscriber::EnvFilter::try_from_default_env()
        .or_else(|_| tracing_subscriber::EnvFilter::try_new("info"))?;

    tracing_subscriber::registry()
        .with(env_filter)
        .with(
            tracing_subscriber::fmt::layer()
                .with_writer(std::io::stderr)
                .with_target(true)
                .with_thread_ids(true)
                .with_file(true)
                .with_line_number(true),
        )
        .with(telemetry)
        .init();

    Ok(())
}

pub fn init_tracing_simple() {
    let env_filter = tracing_subscriber::EnvFilter::try_from_default_env()
        .or_else(|_| tracing_subscriber::EnvFilter::try_new("info"))
        .unwrap();

    tracing_subscriber::registry()
        .with(env_filter)
        .with(
            tracing_subscriber::fmt::layer()
                .with_writer(std::io::stderr)
                .with_target(true)
                .with_thread_ids(true)
                .with_file(true)
                .with_line_number(true),
        )
        .init();
}
