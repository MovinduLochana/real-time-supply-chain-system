package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rtscs/services/notification-service/db"
	"go.uber.org/zap"
)

type HealthHandler struct {
	redis  *db.RedisClient
	logger *zap.Logger
}

func NewHealthHandler(redis *db.RedisClient, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		redis:  redis,
		logger: logger,
	}
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Services  map[string]ServiceStatus `json:"services"`
}

type ServiceStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// Health handles GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services: map[string]ServiceStatus{
			"redis": h.checkRedis(ctx),
		},
	}

	// If any service is down, mark overall as unhealthy
	for _, svc := range response.Services {
		if svc.Status != "healthy" {
			response.Status = "unhealthy"
			break
		}
	}

	statusCode := http.StatusOK
	if response.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (h *HealthHandler) checkRedis(ctx context.Context) ServiceStatus {
	if err := h.redis.Ping(ctx); err != nil {
		h.logger.Error("Redis health check failed", zap.Error(err))
		return ServiceStatus{
			Status: "unhealthy",
			Error:  err.Error(),
		}
	}
	return ServiceStatus{
		Status: "healthy",
	}
}
