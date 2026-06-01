package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/rtscs/services/notification-service/config"
	"github.com/rtscs/services/notification-service/models"
	"go.uber.org/zap"
)

type EmailService struct {
	client *sendgrid.Client
	config *config.Config
	logger *zap.Logger
}

func NewEmailService(cfg *config.Config, logger *zap.Logger) *EmailService {
	client := sendgrid.NewSendClient(cfg.SendGridAPIKey)
	return &EmailService{
		client: client,
		config: cfg,
		logger: logger,
	}
}

func (s *EmailService) SendEmail(ctx context.Context, notif *models.Notification) error {
	if notif.Channel != models.ChannelEmail {
		return fmt.Errorf("invalid channel: expected email, got %s", notif.Channel)
	}

	from := mail.NewEmail("RTSCS", "noreply@rtscs.com")
	to := mail.NewEmail("", notif.RecipientID)
	message := mail.NewSingleEmail(from, notif.Subject, to, notif.Body, "")

	response, err := s.client.SendWithContext(ctx, message)
	if err != nil {
		s.logger.Error("Failed to send email",
			zap.Error(err),
			zap.String("notification_id", notif.ID),
			zap.String("recipient", notif.RecipientID),
			zap.String("correlation_id", notif.CorrelationID),
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		errMsg := fmt.Sprintf("SendGrid API error: status %d, body %s", response.StatusCode, response.Body)
		s.logger.Error("SendGrid API error",
			zap.String("notification_id", notif.ID),
			zap.Int("status_code", response.StatusCode),
			zap.String("response_body", response.Body),
			zap.String("correlation_id", notif.CorrelationID),
		)
		return fmt.Errorf(errMsg)
	}

	s.logger.Info("Email sent successfully",
		zap.String("notification_id", notif.ID),
		zap.String("recipient", notif.RecipientID),
		zap.String("correlation_id", notif.CorrelationID),
	)

	now := time.Now()
	notif.SentAt = &now
	notif.Status = models.NotificationStatusSent

	return nil
}

// Email template helpers

func BuildOrderConfirmationEmail(orderID string, amount float64) (subject, body string) {
	subject = fmt.Sprintf("Order Confirmation - #%s", orderID)
	body = fmt.Sprintf(`
Hello,

Thank you for your order!

Order ID: %s
Amount: $%.2f

Your order has been received and is being processed.

You can track your order status in your account dashboard.

Best regards,
RTSCS Team
`, orderID, amount)
	return
}

func BuildOrderCancellationEmail(orderID, reason string) (subject, body string) {
	subject = fmt.Sprintf("Order Cancelled - #%s", orderID)
	body = fmt.Sprintf(`
Hello,

Your order has been cancelled.

Order ID: %s
Reason: %s

If you have any questions, please contact our support team.

Best regards,
RTSCS Team
`, orderID, reason)
	return
}

func BuildLowStockAlertEmail(productName string, currentStock int) (subject, body string) {
	subject = fmt.Sprintf("Low Stock Alert - %s", productName)
	body = fmt.Sprintf(`
Hello,

We wanted to notify you that one of your favorite products is running low on stock.

Product: %s
Current Stock: %d units

If you're interested in this product, we recommend ordering soon to avoid missing out.

Best regards,
RTSCS Team
`, productName, currentStock)
	return
}
