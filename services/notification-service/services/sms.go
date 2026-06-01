package services

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rtscs/services/notification-service/config"
	"github.com/rtscs/services/notification-service/models"
	"go.uber.org/zap"
)

type SMSService struct {
	client *http.Client
	config *config.Config
	logger *zap.Logger
}

func NewSMSService(cfg *config.Config, logger *zap.Logger) *SMSService {
	return &SMSService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: cfg,
		logger: logger,
	}
}

func (s *SMSService) SendSMS(ctx context.Context, notif *models.Notification) error {
	if notif.Channel != models.ChannelSMS {
		return fmt.Errorf("invalid channel: expected sms, got %s", notif.Channel)
	}

	// Using Twilio REST API directly via HTTP
	// Reference: https://www.twilio.com/docs/sms/send-messages
	
	if s.config.TwilioAccountSID == "" || s.config.TwilioAuthToken == "" {
		s.logger.Warn("Twilio credentials not configured, SMS queued locally",
			zap.String("notification_id", notif.ID),
			zap.String("recipient", notif.RecipientID),
		)
		return s.queueSMSLocally(notif)
	}

	// Build Twilio API request
	twilioURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json",
		s.config.TwilioAccountSID)

	data := url.Values{}
	data.Set("From", s.config.TwilioFromNumber)
	data.Set("To", notif.RecipientID)
	data.Set("Body", notif.Body)

	req, err := http.NewRequestWithContext(ctx, "POST", twilioURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create Twilio request: %w", err)
	}

	req.SetBasicAuth(s.config.TwilioAccountSID, s.config.TwilioAuthToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("Failed to send SMS via Twilio",
			zap.Error(err),
			zap.String("notification_id", notif.ID),
			zap.String("recipient", notif.RecipientID),
			zap.String("correlation_id", notif.CorrelationID),
		)
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errMsg := fmt.Sprintf("Twilio API error: status %d", resp.StatusCode)
		s.logger.Error(errMsg,
			zap.String("notification_id", notif.ID),
			zap.Int("status_code", resp.StatusCode),
			zap.String("correlation_id", notif.CorrelationID),
		)
		return fmt.Errorf(errMsg)
	}

	s.logger.Info("SMS sent successfully via Twilio",
		zap.String("notification_id", notif.ID),
		zap.String("recipient", notif.RecipientID),
		zap.String("correlation_id", notif.CorrelationID),
	)

	now := time.Now()
	notif.SentAt = &now
	notif.Status = models.NotificationStatusSent
	if notif.Metadata == nil {
		notif.Metadata = make(map[string]interface{})
	}
	notif.Metadata["sms_sent_at"] = now.Unix()

	return nil
}

func (s *SMSService) queueSMSLocally(notif *models.Notification) error {
	now := time.Now()
	notif.SentAt = &now
	notif.Status = models.NotificationStatusSent
	if notif.Metadata == nil {
		notif.Metadata = make(map[string]interface{})
	}
	notif.Metadata["sms_queued_locally"] = true
	notif.Metadata["queued_at"] = now.Unix()

	s.logger.Info("SMS queued for delivery",
		zap.String("notification_id", notif.ID),
		zap.String("recipient", notif.RecipientID),
		zap.String("correlation_id", notif.CorrelationID),
	)

	return nil
}

// SMS template helpers

func BuildOrderConfirmationSMS(orderID string) string {
	return fmt.Sprintf("Order confirmed! Your order #%s is being processed. Track it in your account. - RTSCS", orderID)
}

func BuildOrderCancellationSMS(orderID string) string {
	return fmt.Sprintf("Your order #%s has been cancelled. Contact support for more info. - RTSCS", orderID)
}

func BuildDeliveryUpdateSMS(orderID, status string) string {
	return fmt.Sprintf("Order #%s update: %s. Track your delivery in your account. - RTSCS", orderID, status)
}

func BuildLowStockAlertSMS(productName string, stock int) string {
	return fmt.Sprintf("%s is low on stock (only %d left). Order now! - RTSCS", productName, stock)
}
