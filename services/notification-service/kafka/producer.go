package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/rtscs/services/notification-service/config"
	"github.com/rtscs/services/notification-service/models"
	"go.uber.org/zap"
)

type Producer struct {
	writer *kafka.Writer
	config *config.Config
	logger *zap.Logger
	lock   sync.Mutex
}

func NewProducer(cfg *config.Config, logger *zap.Logger) *Producer {
	return &Producer{
		config: cfg,
		logger: logger,
	}
}

func (p *Producer) Start(ctx context.Context) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	writer := &kafka.Writer{
		Addr:         kafka.TCP(p.config.KafkaBrokers...),
		Compression:  kafka.Snappy,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	p.writer = writer
	p.logger.Info("Started Kafka producer",
		zap.Strings("brokers", p.config.KafkaBrokers),
	)

	return nil
}

func (p *Producer) PublishNotificationSentEvent(ctx context.Context, event *models.NotificationSentEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal notification sent event: %w", err)
	}

	return p.publishEvent(ctx, "notification-events", data, event.CorrelationID)
}

func (p *Producer) PublishNotificationDeliveredEvent(ctx context.Context, event *models.NotificationDeliveredEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal notification delivered event: %w", err)
	}

	return p.publishEvent(ctx, "notification-events", data, event.CorrelationID)
}

func (p *Producer) PublishNotificationFailedEvent(ctx context.Context, event *models.NotificationFailedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal notification failed event: %w", err)
	}

	return p.publishEvent(ctx, "notification-events", data, event.CorrelationID)
}

func (p *Producer) publishEvent(ctx context.Context, topic string, data []byte, correlationID string) error {
	p.lock.Lock()
	writer := p.writer
	p.lock.Unlock()

	if writer == nil {
		return fmt.Errorf("producer not started")
	}

	messages := []kafka.Message{
		{
			Topic: topic,
			Value: data,
			Headers: []kafka.Header{
				{
					Key:   "correlation-id",
					Value: []byte(correlationID),
				},
			},
		},
	}

	err := writer.WriteMessages(ctx, messages...)
	if err != nil {
		p.logger.Error("Failed to publish event",
			zap.Error(err),
			zap.String("topic", topic),
			zap.String("correlation_id", correlationID),
		)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.Debug("Published event",
		zap.String("topic", topic),
		zap.String("correlation_id", correlationID),
	)

	return nil
}

func (p *Producer) Stop(ctx context.Context) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.writer == nil {
		return nil
	}

	p.logger.Info("Stopping Kafka producer")
	return p.writer.Close()
}
