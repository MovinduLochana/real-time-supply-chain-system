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

type Consumer struct {
	reader          *kafka.Reader
	config          *config.Config
	logger          *zap.Logger
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	eventHandlers   map[string]EventHandler
	handlersLock    sync.RWMutex
}

type EventHandler func(ctx context.Context, message []byte, correlationID string) error

func NewConsumer(cfg *config.Config, logger *zap.Logger) *Consumer {
	return &Consumer{
		config:        cfg,
		logger:        logger,
		eventHandlers: make(map[string]EventHandler),
	}
}

func (c *Consumer) RegisterHandler(eventType string, handler EventHandler) {
	c.handlersLock.Lock()
	defer c.handlersLock.Unlock()
	c.eventHandlers[eventType] = handler
}

func (c *Consumer) Start(ctx context.Context) error {
	c.ctx, c.cancel = context.WithCancel(ctx)

	// Create reader for all topics
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        c.config.KafkaBrokers,
		GroupID:        c.config.KafkaConsumerGroup,
		StartOffset:    kafka.LastOffset,
		CommitInterval: time.Second,
		MaxBytes:       10e6, // 10MB
		GroupTopics:    c.config.KafkaTopics,
	})

	c.reader = reader

	c.logger.Info("Starting Kafka consumer",
		zap.Strings("brokers", c.config.KafkaBrokers),
		zap.Strings("topics", c.config.KafkaTopics),
		zap.String("group_id", c.config.KafkaConsumerGroup),
	)

	c.wg.Add(1)
	go c.processMessages()

	return nil
}

func (c *Consumer) processMessages() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info("Kafka consumer shutting down")
			return
		default:
		}

		ctx, cancel := context.WithTimeout(c.ctx, time.Duration(c.config.KafkaReadTimeoutSeconds)*time.Second)
		msg, err := c.reader.ReadMessage(ctx)
		cancel()

		if err != nil {
			if err == context.DeadlineExceeded {
				continue
			}
			if err == context.Canceled {
				return
			}
			c.logger.Error("Failed to read message from Kafka",
				zap.Error(err),
				zap.String("topic", msg.Topic),
				zap.Int("partition", msg.Partition),
			)
			continue
		}

		correlationID := ""
		if len(msg.Headers) > 0 {
			for _, header := range msg.Headers {
				if header.Key == "correlation-id" {
					correlationID = string(header.Value)
					break
				}
			}
		}

		c.handleMessage(c.ctx, msg.Topic, msg.Value, correlationID)
	}
}

func (c *Consumer) handleMessage(ctx context.Context, topic string, value []byte, correlationID string) {
	c.handlersLock.RLock()
	handler, exists := c.eventHandlers[topic]
	c.handlersLock.RUnlock()

	if !exists {
		c.logger.Warn("No handler registered for topic",
			zap.String("topic", topic),
			zap.String("correlation_id", correlationID),
		)
		return
	}

	if err := handler(ctx, value, correlationID); err != nil {
		c.logger.Error("Failed to handle message",
			zap.Error(err),
			zap.String("topic", topic),
			zap.String("correlation_id", correlationID),
		)
	}
}

func (c *Consumer) Stop(ctx context.Context) error {
	c.logger.Info("Stopping Kafka consumer")
	c.cancel()

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for consumer to stop: %w", ctx.Err())
	}

	if c.reader != nil {
		return c.reader.Close()
	}

	return nil
}

// Event unmarshaling helpers

func UnmarshalOrderCreatedEvent(data []byte) (*models.OrderCreatedEvent, error) {
	var event models.OrderCreatedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OrderCreatedEvent: %w", err)
	}
	return &event, nil
}

func UnmarshalOrderCancelledEvent(data []byte) (*models.OrderCancelledEvent, error) {
	var event models.OrderCancelledEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OrderCancelledEvent: %w", err)
	}
	return &event, nil
}

func UnmarshalLowStockAlertEvent(data []byte) (*models.LowStockAlertEvent, error) {
	var event models.LowStockAlertEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal LowStockAlertEvent: %w", err)
	}
	return &event, nil
}

func UnmarshalStockUpdatedEvent(data []byte) (*models.StockUpdatedEvent, error) {
	var event models.StockUpdatedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal StockUpdatedEvent: %w", err)
	}
	return &event, nil
}
