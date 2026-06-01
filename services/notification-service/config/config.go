package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// Server
	ServerPort int
	Environment string

	// Kafka
	KafkaBrokers       []string
	KafkaConsumerGroup string
	KafkaTopics        []string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RedisMaxRetries int
	RedisTTL      time.Duration

	// SendGrid
	SendGridAPIKey string

	// Twilio
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string

	// Firebase Cloud Messaging
	FCMProjectID string
	FCMKeyPath   string

	// OpenTelemetry
	OTelEnabled bool
	OTelEndpoint string

	// Logging
	LogLevel string

	// Worker Pool
	WorkerPoolSize int

	// Retry
	MaxRetries       int
	RetryWaitSeconds int

	// Timeouts
	KafkaReadTimeoutSeconds int
	HTTPTimeoutSeconds      int
}

func Load() *Config {
	return &Config{
		ServerPort:  getEnvInt("SERVER_PORT", 3002),
		Environment: getEnv("ENVIRONMENT", "development"),

		KafkaBrokers:       strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		KafkaConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "notification-service"),
		KafkaTopics:        strings.Split(getEnv("KAFKA_TOPICS", "order-events,inventory-events"), ","),

		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:         getEnvInt("REDIS_DB", 0),
		RedisMaxRetries: getEnvInt("REDIS_MAX_RETRIES", 3),
		RedisTTL:        time.Duration(getEnvInt("REDIS_TTL_DAYS", 30)) * 24 * time.Hour,

		SendGridAPIKey: getEnv("SENDGRID_API_KEY", ""),

		TwilioAccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioFromNumber: getEnv("TWILIO_FROM_NUMBER", ""),

		FCMProjectID: getEnv("FCM_PROJECT_ID", ""),
		FCMKeyPath:   getEnv("FCM_KEY_PATH", ""),

		OTelEnabled:  getEnvBool("OTEL_ENABLED", false),
		OTelEndpoint: getEnv("OTEL_ENDPOINT", "localhost:4318"),

		LogLevel:            getEnv("LOG_LEVEL", "info"),
		WorkerPoolSize:      getEnvInt("WORKER_POOL_SIZE", 10),
		MaxRetries:          getEnvInt("MAX_RETRIES", 3),
		RetryWaitSeconds:    getEnvInt("RETRY_WAIT_SECONDS", 2),
		KafkaReadTimeoutSeconds: getEnvInt("KAFKA_READ_TIMEOUT_SECONDS", 30),
		HTTPTimeoutSeconds:      getEnvInt("HTTP_TIMEOUT_SECONDS", 30),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return defaultVal
}
