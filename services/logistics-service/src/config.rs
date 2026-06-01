use std::env;

#[derive(Clone, Debug)]
pub struct Config {
    pub database_url: String,
    pub kafka_brokers: String,
    pub kafka_consumer_group: String,
    pub service_port: u16,
    pub service_name: String,
    #[allow(dead_code)]
    pub rust_log: String,
}

impl Config {
    pub fn from_env() -> Self {
        dotenv::dotenv().ok();

        let database_url = env::var("DATABASE_URL")
            .expect("DATABASE_URL not set");
        
        let kafka_brokers = env::var("KAFKA_BROKERS")
            .expect("KAFKA_BROKERS not set");
        
        let kafka_consumer_group = env::var("KAFKA_CONSUMER_GROUP")
            .unwrap_or_else(|_| "logistics-service".to_string());
        
        let service_port: u16 = env::var("SERVICE_PORT")
            .unwrap_or_else(|_| "3001".to_string())
            .parse()
            .expect("SERVICE_PORT must be a valid u16");
        
        let service_name = env::var("SERVICE_NAME")
            .unwrap_or_else(|_| "logistics-service".to_string());
        
        let rust_log = env::var("RUST_LOG")
            .unwrap_or_else(|_| "info".to_string());

        Config {
            database_url,
            kafka_brokers,
            kafka_consumer_group,
            service_port,
            service_name,
            rust_log,
        }
    }

    pub fn kafka_brokers_vec(&self) -> Vec<&str> {
        self.kafka_brokers.split(',').map(|s| s.trim()).collect()
    }
}
