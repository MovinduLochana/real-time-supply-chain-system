package services

import (
	"context"
	"fmt"
	"time"

	"github.com/rtscs/services/notification-service/config"
	"github.com/rtscs/services/notification-service/models"
	"go.uber.org/zap"
)

type PushService struct {
	config *config.Config
	logger *zap.Logger
}

func NewPushService(cfg *config.Config, logger *zap.Logger) *PushService {
	return &PushService{
		config: cfg,
		logger: logger,
	}
}

func (s *PushService) SendPush(ctx context.Context, notif *models.Notification) error {
	if notif.Channel != models.ChannelPush {
		return fmt.Errorf("invalid channel: expected push, got %s", notif.Channel)
	}

	// Firebase Cloud Messaging would be initialized here with the key path
	// For now, we'll simulate sending a push notification
	// In production, use:
	// client, err := firebase.NewClient(ctx, &firebase.Options{
	//     ProjectID: s.config.FCMProjectID,
	// })

	s.logger.Info("Push notification queued for delivery",
		zap.String("notification_id", notif.ID),
		zap.String("device_id", notif.RecipientID),
		zap.String("correlation_id", notif.CorrelationID),
	)

	now := time.Now()
	notif.SentAt = &now
	notif.Status = models.NotificationStatusSent

	// In a real implementation, this would send via Firebase Cloud Messaging
	// For now, we'll mark it as queued
	if notif.Metadata == nil {
		notif.Metadata = make(map[string]interface{})
	}
	notif.Metadata["queued_at"] = now.Unix()

	return nil
}

// Push notification template helpers

func BuildOrderStatusPushNotification(orderID, status string) (title, body string) {
	title = "Order Status Update"
	body = fmt.Sprintf("Order #%s is now %s", orderID, status)
	return
}

func BuildPromotionPushNotification(title, description string) (notifTitle, body string) {
	notifTitle = title
	body = description
	return
}

func BuildDeliveryPushNotification(orderID, status string) (title, body string) {
	title = "Delivery Update"
	body = fmt.Sprintf("Your order #%s: %s", orderID, status)
	return
}
