package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rtscs/services/notification-service/config"
	"github.com/rtscs/services/notification-service/models"
	"go.uber.org/zap"
)

type RedisClient struct {
	client *redis.Client
	config *config.Config
	logger *zap.Logger
}

func NewRedisClient(cfg *config.Config, logger *zap.Logger) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:       cfg.RedisAddr,
		Password:   cfg.RedisPassword,
		DB:         cfg.RedisDB,
		MaxRetries: cfg.RedisMaxRetries,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis", zap.String("addr", cfg.RedisAddr))

	return &RedisClient{
		client: client,
		config: cfg,
		logger: logger,
	}, nil
}

// Notification methods

func (r *RedisClient) SaveNotification(ctx context.Context, notif *models.Notification) error {
	data, err := json.Marshal(notif)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	key := fmt.Sprintf("notification:%s", notif.ID)
	if err := r.client.Set(ctx, key, data, r.config.RedisTTL).Err(); err != nil {
		return fmt.Errorf("failed to save notification to Redis: %w", err)
	}

	return nil
}

func (r *RedisClient) GetNotification(ctx context.Context, id string) (*models.Notification, error) {
	key := fmt.Sprintf("notification:%s", id)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("notification not found")
		}
		return nil, fmt.Errorf("failed to get notification from Redis: %w", err)
	}

	var notif models.Notification
	if err := json.Unmarshal([]byte(data), &notif); err != nil {
		return nil, fmt.Errorf("failed to unmarshal notification: %w", err)
	}

	return &notif, nil
}

func (r *RedisClient) UpdateNotification(ctx context.Context, notif *models.Notification) error {
	data, err := json.Marshal(notif)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	key := fmt.Sprintf("notification:%s", notif.ID)
	ttl := r.client.TTL(ctx, key).Val()
	if ttl == -1 || ttl == -2 {
		ttl = r.config.RedisTTL
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to update notification in Redis: %w", err)
	}

	return nil
}

func (r *RedisClient) DeleteNotification(ctx context.Context, id string) error {
	key := fmt.Sprintf("notification:%s", id)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete notification from Redis: %w", err)
	}
	return nil
}

// User preferences methods

func (r *RedisClient) SaveUserPreferences(ctx context.Context, prefs *models.UserPreferences) error {
	data, err := json.Marshal(prefs)
	if err != nil {
		return fmt.Errorf("failed to marshal user preferences: %w", err)
	}

	key := fmt.Sprintf("user-preferences:%s", prefs.UserID)
	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to save user preferences to Redis: %w", err)
	}

	return nil
}

func (r *RedisClient) GetUserPreferences(ctx context.Context, userID string) (*models.UserPreferences, error) {
	key := fmt.Sprintf("user-preferences:%s", userID)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Return default preferences if not found
			return &models.UserPreferences{
				UserID:      userID,
				EmailOptIn:  true,
				SMSOptIn:    true,
				PushOptIn:   true,
				DoNotDisturb: false,
				Channels:    make(map[string]string),
				UpdatedAt:   time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get user preferences from Redis: %w", err)
	}

	var prefs models.UserPreferences
	if err := json.Unmarshal([]byte(data), &prefs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user preferences: %w", err)
	}

	return &prefs, nil
}

// Queue methods

func (r *RedisClient) EnqueueNotification(ctx context.Context, channel string, notifID string) error {
	key := fmt.Sprintf("notification-queue:%s", channel)
	if err := r.client.RPush(ctx, key, notifID).Err(); err != nil {
		return fmt.Errorf("failed to enqueue notification: %w", err)
	}
	return nil
}

func (r *RedisClient) DequeueNotification(ctx context.Context, channel string) (string, error) {
	key := fmt.Sprintf("notification-queue:%s", channel)
	result, err := r.client.LPop(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", fmt.Errorf("failed to dequeue notification: %w", err)
	}
	return result, nil
}

func (r *RedisClient) GetQueueLength(ctx context.Context, channel string) (int64, error) {
	key := fmt.Sprintf("notification-queue:%s", channel)
	length, err := r.client.LLen(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get queue length: %w", err)
	}
	return length, nil
}

// Health check

func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
