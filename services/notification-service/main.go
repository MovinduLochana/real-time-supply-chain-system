package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rtscs/services/notification-service/api"
	"github.com/rtscs/services/notification-service/config"
	"github.com/rtscs/services/notification-service/db"
	"github.com/rtscs/services/notification-service/kafka"
	"github.com/rtscs/services/notification-service/models"
	"github.com/rtscs/services/notification-service/services"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"
)

func main() {
	// Load environment variables
	_ = godotenv.Load()

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Info("Starting Notification Service")

	// Load configuration
	cfg := config.Load()
	logger.Info("Configuration loaded",
		zap.String("environment", cfg.Environment),
		zap.Int("server_port", cfg.ServerPort),
		zap.String("redis_addr", cfg.RedisAddr),
	)

	// Initialize OpenTelemetry if enabled
	if cfg.OTelEnabled {
		initTracing(logger)
	}

	// Initialize Redis client
	redisClient, err := db.NewRedisClient(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Redis client", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize Kafka producer
	producer := kafka.NewProducer(cfg, logger)
	if err := producer.Start(context.Background()); err != nil {
		logger.Fatal("Failed to start Kafka producer", zap.Error(err))
	}
	defer producer.Stop(context.Background())

	// Initialize Kafka consumer
	consumer := kafka.NewConsumer(cfg, logger)

	// Initialize notification service
	notifSvc := services.NewNotificationService(redisClient, producer, cfg, logger)
	if err := notifSvc.Start(context.Background()); err != nil {
		logger.Fatal("Failed to start notification service", zap.Error(err))
	}
	defer notifSvc.Stop(context.Background())

	// Register Kafka event handlers
	registerEventHandlers(consumer, notifSvc, logger)

	// Start Kafka consumer
	if err := consumer.Start(context.Background()); err != nil {
		logger.Fatal("Failed to start Kafka consumer", zap.Error(err))
	}
	defer consumer.Stop(context.Background())

	// Initialize HTTP router
	router := api.NewRouter(notifSvc, redisClient, logger)

	// Start HTTP server
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.HTTPTimeoutSeconds) * time.Second,
		WriteTimeout:   time.Duration(cfg.HTTPTimeoutSeconds) * time.Second,
		IdleTimeout:    30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	go func() {
		logger.Info("Starting HTTP server", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutdown signal received, graceful shutdown starting")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	logger.Info("Notification Service stopped")
}

func registerEventHandlers(
	consumer *kafka.Consumer,
	notifSvc *services.NotificationService,
	logger *zap.Logger,
) {
	// Order created event handler
	consumer.RegisterHandler("order-events", func(ctx context.Context, message []byte, correlationID string) error {
		// Try to unmarshal as different event types
		if event, err := kafka.UnmarshalOrderCreatedEvent(message); err == nil {
			logger.Info("Processing OrderCreatedEvent",
				zap.String("order_id", event.OrderID),
				zap.String("user_id", event.UserID),
				zap.String("correlation_id", correlationID),
			)

			// Create email notification
			subject, body := services.BuildOrderConfirmationEmail(event.OrderID, event.OrderAmount)
			notif := &models.Notification{
				UserID:        event.UserID,
				Channel:       models.ChannelEmail,
				Subject:       subject,
				Body:          body,
				CorrelationID: correlationID,
				Metadata: map[string]interface{}{
					"event_type": "order_created",
					"order_id":   event.OrderID,
					"amount":     event.OrderAmount,
				},
			}
			return notifSvc.CreateNotification(ctx, notif)
		}

		if event, err := kafka.UnmarshalOrderCancelledEvent(message); err == nil {
			logger.Info("Processing OrderCancelledEvent",
				zap.String("order_id", event.OrderID),
				zap.String("user_id", event.UserID),
				zap.String("correlation_id", correlationID),
			)

			// Create email notification
			subject, body := services.BuildOrderCancellationEmail(event.OrderID, event.CancelReason)
			notif := &models.Notification{
				UserID:        event.UserID,
				Channel:       models.ChannelEmail,
				Subject:       subject,
				Body:          body,
				CorrelationID: correlationID,
				Metadata: map[string]interface{}{
					"event_type": "order_cancelled",
					"order_id":   event.OrderID,
					"reason":     event.CancelReason,
				},
			}
			return notifSvc.CreateNotification(ctx, notif)
		}

		logger.Warn("Unable to unmarshal order event", zap.String("correlation_id", correlationID))
		return nil
	})

	// Low stock alert event handler
	consumer.RegisterHandler("inventory-events", func(ctx context.Context, message []byte, correlationID string) error {
		if event, err := kafka.UnmarshalLowStockAlertEvent(message); err == nil {
			logger.Info("Processing LowStockAlertEvent",
				zap.String("product_id", event.ProductID),
				zap.String("correlation_id", correlationID),
			)

			// Create email notification for low stock
			subject, body := services.BuildLowStockAlertEmail(event.ProductName, event.CurrentStock)
			notif := &models.Notification{
				UserID:        "admin", // Would come from product owner mapping
				Channel:       models.ChannelEmail,
				Subject:       subject,
				Body:          body,
				CorrelationID: correlationID,
				Metadata: map[string]interface{}{
					"event_type": "low_stock_alert",
					"product_id": event.ProductID,
					"stock":      event.CurrentStock,
				},
			}
			return notifSvc.CreateNotification(ctx, notif)
		}

		logger.Warn("Unable to unmarshal inventory event", zap.String("correlation_id", correlationID))
		return nil
	})
}

func initTracing(logger *zap.Logger) {
	// Create a resource to describe the service
	resource, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("notification-service"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		logger.Error("Failed to create resource", zap.Error(err))
		return
	}

	// Create a basic trace provider (without an exporter for now)
	// In production, you would configure an actual exporter like Jaeger, Tempo, etc.
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
	)

	otel.SetTracerProvider(tracerProvider)
	logger.Info("OpenTelemetry tracing initialized")
}
