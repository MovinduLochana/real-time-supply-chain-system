package models

import "time"

type NotificationStatus string

const (
	NotificationStatusPending    NotificationStatus = "PENDING"
	NotificationStatusSent       NotificationStatus = "SENT"
	NotificationStatusDelivered  NotificationStatus = "DELIVERED"
	NotificationStatusFailed     NotificationStatus = "FAILED"
	NotificationStatusUnsubscribed NotificationStatus = "UNSUBSCRIBED"
)

type NotificationChannel string

const (
	ChannelEmail NotificationChannel = "email"
	ChannelSMS   NotificationChannel = "sms"
	ChannelPush  NotificationChannel = "push"
)

type Notification struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	Channel      NotificationChannel    `json:"channel"`
	RecipientID  string                 `json:"recipient_id"`      // Email, phone, or device ID
	Subject      string                 `json:"subject"`
	Body         string                 `json:"body"`
	Status       NotificationStatus     `json:"status"`
	Metadata     map[string]interface{} `json:"metadata"`
	RetryCount   int                    `json:"retry_count"`
	MaxRetries   int                    `json:"max_retries"`
	CreatedAt    time.Time              `json:"created_at"`
	SentAt       *time.Time             `json:"sent_at"`
	DeliveredAt  *time.Time             `json:"delivered_at"`
	FailedAt     *time.Time             `json:"failed_at"`
	ErrorMessage string                 `json:"error_message"`
	CorrelationID string                `json:"correlation_id"`
}

type UserPreferences struct {
	UserID          string            `json:"user_id"`
	EmailOptIn      bool              `json:"email_opt_in"`
	SMSOptIn        bool              `json:"sms_opt_in"`
	PushOptIn       bool              `json:"push_opt_in"`
	DoNotDisturb    bool              `json:"do_not_disturb"`
	DoNotDisturbEnd *time.Time        `json:"do_not_disturb_end"`
	Channels        map[string]string `json:"channels"` // channel -> recipient (email, phone, device_id)
	UpdatedAt       time.Time         `json:"updated_at"`
}

// Event models for Kafka

type OrderCreatedEvent struct {
	OrderID      string `json:"order_id"`
	UserID       string `json:"user_id"`
	OrderAmount  float64 `json:"order_amount"`
	CreatedAt    string `json:"created_at"`
	CorrelationID string `json:"correlation_id"`
}

type OrderCancelledEvent struct {
	OrderID       string `json:"order_id"`
	UserID        string `json:"user_id"`
	CancelReason  string `json:"cancel_reason"`
	CancelledAt   string `json:"cancelled_at"`
	CorrelationID string `json:"correlation_id"`
}

type LowStockAlertEvent struct {
	ProductID     string `json:"product_id"`
	ProductName   string `json:"product_name"`
	CurrentStock  int    `json:"current_stock"`
	AlertedAt     string `json:"alerted_at"`
	CorrelationID string `json:"correlation_id"`
}

type StockUpdatedEvent struct {
	ProductID     string `json:"product_id"`
	ProductName   string `json:"product_name"`
	PreviousStock int    `json:"previous_stock"`
	NewStock      int    `json:"new_stock"`
	UpdatedAt     string `json:"updated_at"`
	CorrelationID string `json:"correlation_id"`
}

// Notification event models for Kafka producer

type NotificationSentEvent struct {
	NotificationID string `json:"notification_id"`
	UserID        string `json:"user_id"`
	Channel       string `json:"channel"`
	SentAt        string `json:"sent_at"`
	CorrelationID string `json:"correlation_id"`
}

type NotificationDeliveredEvent struct {
	NotificationID string `json:"notification_id"`
	UserID        string `json:"user_id"`
	Channel       string `json:"channel"`
	DeliveredAt   string `json:"delivered_at"`
	CorrelationID string `json:"correlation_id"`
}

type NotificationFailedEvent struct {
	NotificationID string `json:"notification_id"`
	UserID        string `json:"user_id"`
	Channel       string `json:"channel"`
	ErrorMessage  string `json:"error_message"`
	FailedAt      string `json:"failed_at"`
	CorrelationID string `json:"correlation_id"`
}
