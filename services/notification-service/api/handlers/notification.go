package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rtscs/services/notification-service/middleware"
	"github.com/rtscs/services/notification-service/models"
	"github.com/rtscs/services/notification-service/services"
	"go.uber.org/zap"
)

type NotificationHandler struct {
	svc    *services.NotificationService
	logger *zap.Logger
}

func NewNotificationHandler(svc *services.NotificationService, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		svc:    svc,
		logger: logger,
	}
}

// SendEmailRequest represents an email notification request
type SendEmailRequest struct {
	UserID       string            `json:"user_id"`
	RecipientID  string            `json:"recipient_id"`
	Subject      string            `json:"subject"`
	Body         string            `json:"body"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// SendEmail handles POST /notifications/email
func (h *NotificationHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := middleware.GetCorrelationID(ctx)

	var req SendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid request body",
			zap.Error(err),
			zap.String("correlation_id", correlationID),
		)
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.UserID == "" || req.RecipientID == "" || req.Subject == "" || req.Body == "" {
		h.logger.Warn("Missing required fields",
			zap.String("correlation_id", correlationID),
		)
		http.Error(w, `{"error": "missing required fields"}`, http.StatusBadRequest)
		return
	}

	notif := &models.Notification{
		UserID:       req.UserID,
		Channel:      models.ChannelEmail,
		RecipientID:  req.RecipientID,
		Subject:      req.Subject,
		Body:         req.Body,
		Metadata:     req.Metadata,
		CorrelationID: correlationID,
	}

	if err := h.svc.CreateNotification(ctx, notif); err != nil {
		h.logger.Error("Failed to create email notification",
			zap.Error(err),
			zap.String("user_id", req.UserID),
			zap.String("correlation_id", correlationID),
		)
		http.Error(w, `{"error": "failed to create notification"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"notification_id": notif.ID,
		"status":          "PENDING",
	})
}

// SendSMSRequest represents an SMS notification request
type SendSMSRequest struct {
	UserID       string            `json:"user_id"`
	RecipientID  string            `json:"recipient_id"`
	Body         string            `json:"body"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// SendSMS handles POST /notifications/sms
func (h *NotificationHandler) SendSMS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := middleware.GetCorrelationID(ctx)

	var req SendSMSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid request body",
			zap.Error(err),
			zap.String("correlation_id", correlationID),
		)
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.UserID == "" || req.RecipientID == "" || req.Body == "" {
		h.logger.Warn("Missing required fields",
			zap.String("correlation_id", correlationID),
		)
		http.Error(w, `{"error": "missing required fields"}`, http.StatusBadRequest)
		return
	}

	notif := &models.Notification{
		UserID:       req.UserID,
		Channel:      models.ChannelSMS,
		RecipientID:  req.RecipientID,
		Body:         req.Body,
		Metadata:     req.Metadata,
		CorrelationID: correlationID,
	}

	if err := h.svc.CreateNotification(ctx, notif); err != nil {
		h.logger.Error("Failed to create SMS notification",
			zap.Error(err),
			zap.String("user_id", req.UserID),
			zap.String("correlation_id", correlationID),
		)
		http.Error(w, `{"error": "failed to create notification"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"notification_id": notif.ID,
		"status":          "PENDING",
	})
}

// SendPushRequest represents a push notification request
type SendPushRequest struct {
	UserID       string            `json:"user_id"`
	RecipientID  string            `json:"recipient_id"`
	Subject      string            `json:"subject"`
	Body         string            `json:"body"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// SendPush handles POST /notifications/push
func (h *NotificationHandler) SendPush(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := middleware.GetCorrelationID(ctx)

	var req SendPushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid request body",
			zap.Error(err),
			zap.String("correlation_id", correlationID),
		)
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.UserID == "" || req.RecipientID == "" || req.Body == "" {
		h.logger.Warn("Missing required fields",
			zap.String("correlation_id", correlationID),
		)
		http.Error(w, `{"error": "missing required fields"}`, http.StatusBadRequest)
		return
	}

	notif := &models.Notification{
		UserID:       req.UserID,
		Channel:      models.ChannelPush,
		RecipientID:  req.RecipientID,
		Subject:      req.Subject,
		Body:         req.Body,
		Metadata:     req.Metadata,
		CorrelationID: correlationID,
	}

	if err := h.svc.CreateNotification(ctx, notif); err != nil {
		h.logger.Error("Failed to create push notification",
			zap.Error(err),
			zap.String("user_id", req.UserID),
			zap.String("correlation_id", correlationID),
		)
		http.Error(w, `{"error": "failed to create notification"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"notification_id": notif.ID,
		"status":          "PENDING",
	})
}

// GetNotificationStatus handles GET /notifications/{id}
func (h *NotificationHandler) GetNotificationStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	correlationID := middleware.GetCorrelationID(ctx)

	if id == "" {
		h.logger.Warn("Missing notification ID",
			zap.String("correlation_id", correlationID),
		)
		http.Error(w, `{"error": "notification id required"}`, http.StatusBadRequest)
		return
	}

	notif, err := h.svc.GetNotification(ctx, id)
	if err != nil {
		h.logger.Warn("Notification not found",
			zap.Error(err),
			zap.String("notification_id", id),
			zap.String("correlation_id", correlationID),
		)
		http.Error(w, `{"error": "notification not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notif)
}

// SetUserPreferencesRequest represents user notification preferences
type SetUserPreferencesRequest struct {
	EmailOptIn  bool              `json:"email_opt_in"`
	SMSOptIn    bool              `json:"sms_opt_in"`
	PushOptIn   bool              `json:"push_opt_in"`
	DoNotDisturb bool             `json:"do_not_disturb"`
	DoNotDisturbEnd *time.Time    `json:"do_not_disturb_end"`
	Channels    map[string]string `json:"channels"`
}

// SetUserPreferences handles POST /notifications/preferences/{user_id}
func (h *NotificationHandler) SetUserPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.PathValue("user_id")
	correlationID := middleware.GetCorrelationID(ctx)

	if userID == "" {
		h.logger.Warn("Missing user ID",
			zap.String("correlation_id", correlationID),
		)
		http.Error(w, `{"error": "user id required"}`, http.StatusBadRequest)
		return
	}

	var req SetUserPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid request body",
			zap.Error(err),
			zap.String("correlation_id", correlationID),
		)
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	prefs := &models.UserPreferences{
		UserID:           userID,
		EmailOptIn:       req.EmailOptIn,
		SMSOptIn:         req.SMSOptIn,
		PushOptIn:        req.PushOptIn,
		DoNotDisturb:     req.DoNotDisturb,
		DoNotDisturbEnd:  req.DoNotDisturbEnd,
		Channels:         req.Channels,
		UpdatedAt:        time.Now(),
	}

	if err := h.svc.SetUserPreferences(ctx, prefs); err != nil {
		h.logger.Error("Failed to set user preferences",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("correlation_id", correlationID),
		)
		http.Error(w, `{"error": "failed to update preferences"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"message": "preferences updated",
	})
}
