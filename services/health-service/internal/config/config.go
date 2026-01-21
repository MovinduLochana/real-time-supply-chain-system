package config

import "os"

type Config struct {
	Port         string
	PostgresURL  string
	KafkaBrokers string
	RedisURL     string
	OtelEndpoint string
	Services     []ServiceConfig
}

type ServiceConfig struct {
	Name    string
	URL     string
	Timeout int
}

func Load() *Config {
	return &Config{
		Port:         getEnv("SERVER_PORT", "8091"),
		PostgresURL:  getEnv("POSTGRES_URL", "postgres://logistics:logistics_secret@localhost:5432/logistics"),
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		RedisURL:     getEnv("REDIS_URL", "localhost:6379"),
		OtelEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		Services: []ServiceConfig{
			{Name: "auth-service", URL: getEnv("AUTH_SERVICE_URL", "http://localhost:8081"), Timeout: 5},
			{Name: "order-service", URL: getEnv("ORDER_SERVICE_URL", "http://localhost:8082"), Timeout: 5},
			{Name: "warehouse-service", URL: getEnv("WAREHOUSE_SERVICE_URL", "http://localhost:8083"), Timeout: 5},
			{Name: "config-service", URL: getEnv("CONFIG_SERVICE_URL", "http://localhost:8090"), Timeout: 5},
			{Name: "api-gateway", URL: getEnv("API_GATEWAY_URL", "http://localhost:8080"), Timeout: 5},
			{Name: "gps-ingestion", URL: getEnv("GPS_INGESTION_URL", "http://localhost:8100"), Timeout: 5},
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
