package services

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/rtscs/services/notification-service/config"
	"github.com/rtscs/services/notification-service/db"
	"github.com/rtscs/services/notification-service/kafka"
	"github.com/rtscs/services/notification-service/models"
	"go.uber.org/zap"
)

type NotificationService struct {
	redis     *db.RedisClient
	producer  *kafka.Producer
	emailSvc  *EmailService
	smsSvc    *SMSService
	pushSvc   *PushService
	config    *config.Config
	logger    *zap.Logger
	workerCh  chan WorkerTask
	wg        sync.WaitGroup
	running   int32
	ctx       context.Context
	cancel    context.CancelFunc
}

type WorkerTask struct {
	NotificationID string
	Retry          bool
}

func NewNotificationService(
	redisClient *db.RedisClient,
	producer *kafka.Producer,
	cfg *config.Config,
	logger *zap.Logger,
) *NotificationService {
	return &NotificationService{
		redis:    redisClient,
		producer: producer,
		emailSvc: NewEmailService(cfg, logger),
		smsSvc:   NewSMSService(cfg, logger),
		pushSvc:  NewPushService(cfg, logger),
		config:   cfg,
		logger:   logger,
		workerCh: make(chan WorkerTask, cfg.WorkerPoolSize*10),
	}
}

func (s *NotificationService) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)
	atomic.StoreInt32(&s.running, 1)

	// Start worker pool
	for i := 0; i < s.config.WorkerPoolSize; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	s.logger.Info("Notification service started",
		zap.Int("worker_pool_size", s.config.WorkerPoolSize),
	)

	return nil
}

func (s *NotificationService) worker(id int) {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Debug("Worker shutting down", zap.Int("worker_id", id))
			return
		case task := <-s.workerCh:
			s.processNotification(s.ctx, task)
		}
	}
}

func (s *NotificationService) CreateNotification(ctx context.Context, notif *models.Notification) error {
	if notif.ID == "" {
		notif.ID = uuid.New().String()
	}

	notif.Status = models.NotificationStatusPending
	notif.CreatedAt = time.Now()
	notif.RetryCount = 0
	notif.MaxRetries = s.config.MaxRetries

	// Check user preferences
	prefs, err := s.redis.GetUserPreferences(ctx, notif.UserID)
	if err != nil {
		s.logger.Warn("Failed to get user preferences",
			zap.Error(err),
			zap.String("user_id", notif.UserID),
		)
		// Continue anyway with defaults
		prefs = &models.UserPreferences{
			UserID:      notif.UserID,
			EmailOptIn:  true,
			SMSOptIn:    true,
			PushOptIn:   true,
			DoNotDisturb: false,
		}
	}

	// Check if user opted out or is in do not disturb
	if !s.isChannelEnabled(prefs, notif.Channel) {
		notif.Status = models.NotificationStatusUnsubscribed
		notif.ErrorMessage = "User has opted out of this notification channel"
		s.logger.Info("Notification skipped - user opted out",
			zap.String("notification_id", notif.ID),
			zap.String("channel", string(notif.Channel)),
			zap.String("user_id", notif.UserID),
		)
		return s.redis.SaveNotification(ctx, notif)
	}

	if prefs.DoNotDisturb && (prefs.DoNotDisturbEnd == nil || time.Now().Before(*prefs.DoNotDisturbEnd)) {
		notif.Status = models.NotificationStatusPending
		s.logger.Info("Notification queued - user in do not disturb",
			zap.String("notification_id", notif.ID),
			zap.String("user_id", notif.UserID),
		)
	}

	// Get recipient from preferences if not specified
	if notif.RecipientID == "" {
		if recipientID, exists := prefs.Channels[string(notif.Channel)]; exists {
			notif.RecipientID = recipientID
		}
	}

	if notif.RecipientID == "" {
		notif.Status = models.NotificationStatusFailed
		notif.ErrorMessage = fmt.Sprintf("No recipient ID found for channel %s", notif.Channel)
		s.logger.Warn("Cannot send notification - no recipient",
			zap.String("notification_id", notif.ID),
			zap.String("channel", string(notif.Channel)),
		)
		return s.redis.SaveNotification(ctx, notif)
	}

	// Save notification
	if err := s.redis.SaveNotification(ctx, notif); err != nil {
		s.logger.Error("Failed to save notification",
			zap.Error(err),
			zap.String("notification_id", notif.ID),
		)
		return err
	}

	// Enqueue for processing
	if notif.Status != models.NotificationStatusUnsubscribed {
		if err := s.redis.EnqueueNotification(ctx, string(notif.Channel), notif.ID); err != nil {
			s.logger.Error("Failed to enqueue notification",
				zap.Error(err),
				zap.String("notification_id", notif.ID),
			)
			return err
		}

		// Submit to worker pool
		s.workerCh <- WorkerTask{
			NotificationID: notif.ID,
			Retry:          false,
		}
	}

	s.logger.Info("Notification created and queued",
		zap.String("notification_id", notif.ID),
		zap.String("user_id", notif.UserID),
		zap.String("channel", string(notif.Channel)),
	)

	return nil
}

func (s *NotificationService) processNotification(ctx context.Context, task WorkerTask) {
	notif, err := s.redis.GetNotification(ctx, task.NotificationID)
	if err != nil {
		s.logger.Error("Failed to get notification for processing",
			zap.Error(err),
			zap.String("notification_id", task.NotificationID),
		)
		return
	}

	// Send notification via appropriate channel
	var sendErr error
	switch notif.Channel {
	case models.ChannelEmail:
		sendErr = s.emailSvc.SendEmail(ctx, notif)
	case models.ChannelSMS:
		sendErr = s.smsSvc.SendSMS(ctx, notif)
	case models.ChannelPush:
		sendErr = s.pushSvc.SendPush(ctx, notif)
	default:
		sendErr = fmt.Errorf("unknown channel: %s", notif.Channel)
	}

	if sendErr != nil {
		notif.RetryCount++
		notif.ErrorMessage = sendErr.Error()

		if notif.RetryCount <= notif.MaxRetries {
			// Calculate exponential backoff
			backoffSeconds := int(math.Pow(2, float64(notif.RetryCount))) * s.config.RetryWaitSeconds
			s.logger.Info("Notification send failed, will retry",
				zap.String("notification_id", notif.ID),
				zap.Int("retry_count", notif.RetryCount),
				zap.Int("max_retries", notif.MaxRetries),
				zap.Int("backoff_seconds", backoffSeconds),
				zap.Error(sendErr),
			)

			// Schedule retry
			time.AfterFunc(time.Duration(backoffSeconds)*time.Second, func() {
				select {
				case <-s.ctx.Done():
					return
				case s.workerCh <- WorkerTask{
					NotificationID: notif.ID,
					Retry:          true,
				}:
				}
			})
		} else {
			// Max retries exceeded
			notif.Status = models.NotificationStatusFailed
			now := time.Now()
			notif.FailedAt = &now

			s.logger.Error("Notification failed after max retries",
				zap.String("notification_id", notif.ID),
				zap.Int("retry_count", notif.RetryCount),
				zap.Error(sendErr),
			)

			// Publish failure event
			failedEvent := &models.NotificationFailedEvent{
				NotificationID: notif.ID,
				UserID:        notif.UserID,
				Channel:       string(notif.Channel),
				ErrorMessage:  sendErr.Error(),
				FailedAt:      time.Now().Format(time.RFC3339),
				CorrelationID: notif.CorrelationID,
			}
			s.publishFailedEvent(ctx, failedEvent)
		}
	} else {
		// Success
		notif.Status = models.NotificationStatusSent
		s.logger.Info("Notification sent successfully",
			zap.String("notification_id", notif.ID),
			zap.String("user_id", notif.UserID),
			zap.String("channel", string(notif.Channel)),
		)

		// Publish sent event
		sentEvent := &models.NotificationSentEvent{
			NotificationID: notif.ID,
			UserID:        notif.UserID,
			Channel:       string(notif.Channel),
			SentAt:        time.Now().Format(time.RFC3339),
			CorrelationID: notif.CorrelationID,
		}
		s.publishSentEvent(ctx, sentEvent)
	}

	// Update notification
	if err := s.redis.UpdateNotification(ctx, notif); err != nil {
		s.logger.Error("Failed to update notification",
			zap.Error(err),
			zap.String("notification_id", notif.ID),
		)
	}
}

func (s *NotificationService) isChannelEnabled(prefs *models.UserPreferences, channel models.NotificationChannel) bool {
	switch channel {
	case models.ChannelEmail:
		return prefs.EmailOptIn
	case models.ChannelSMS:
		return prefs.SMSOptIn
	case models.ChannelPush:
		return prefs.PushOptIn
	default:
		return false
	}
}

func (s *NotificationService) GetNotification(ctx context.Context, id string) (*models.Notification, error) {
	return s.redis.GetNotification(ctx, id)
}

func (s *NotificationService) SetUserPreferences(ctx context.Context, prefs *models.UserPreferences) error {
	prefs.UpdatedAt = time.Now()
	return s.redis.SaveUserPreferences(ctx, prefs)
}

func (s *NotificationService) publishSentEvent(ctx context.Context, event *models.NotificationSentEvent) {
	if err := s.producer.PublishNotificationSentEvent(ctx, event); err != nil {
		s.logger.Error("Failed to publish notification sent event",
			zap.Error(err),
			zap.String("notification_id", event.NotificationID),
		)
	}
}

func (s *NotificationService) publishFailedEvent(ctx context.Context, event *models.NotificationFailedEvent) {
	if err := s.producer.PublishNotificationFailedEvent(ctx, event); err != nil {
		s.logger.Error("Failed to publish notification failed event",
			zap.Error(err),
			zap.String("notification_id", event.NotificationID),
		)
	}
}

func (s *NotificationService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping notification service")
	atomic.StoreInt32(&s.running, 0)
	s.cancel()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		close(s.workerCh)
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for notification service to stop: %w", ctx.Err())
	}
}
