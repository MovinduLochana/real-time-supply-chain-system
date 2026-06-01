package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/rtscs/services/notification-service/api/handlers"
	"github.com/rtscs/services/notification-service/db"
	"github.com/rtscs/services/notification-service/middleware"
	"github.com/rtscs/services/notification-service/services"
	"go.uber.org/zap"
)

func NewRouter(
	notifSvc *services.NotificationService,
	redisClient *db.RedisClient,
	logger *zap.Logger,
) *chi.Mux {
	router := chi.NewRouter()

	// Global middleware
	loggingMW := middleware.NewLoggingMiddleware(logger)
	recoveryMW := middleware.NewRecoveryMiddleware(logger)

	router.Use(loggingMW.Handler)
	router.Use(recoveryMW.Handler)

	// Handlers
	notifHandler := handlers.NewNotificationHandler(notifSvc, logger)
	healthHandler := handlers.NewHealthHandler(redisClient, logger)

	// Health check endpoint
	router.Get("/health", healthHandler.Health)

	// Notification endpoints
	router.Post("/notifications/email", notifHandler.SendEmail)
	router.Post("/notifications/sms", notifHandler.SendSMS)
	router.Post("/notifications/push", notifHandler.SendPush)
	router.Get("/notifications/{id}", notifHandler.GetNotificationStatus)
	router.Post("/notifications/preferences/{user_id}", notifHandler.SetUserPreferences)

	return router
}
