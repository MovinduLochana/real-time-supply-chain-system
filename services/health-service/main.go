package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/joho/godotenv"

	"health-service/internal/config"
	"health-service/internal/handler"
	"health-service/internal/telemetry"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()

	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	log := zapLogger.Sugar()

	cfg := config.Load()

	shutdown, err := telemetry.InitTracer(cfg.OtelEndpoint, "health-service")
	if err != nil {
		log.Warnf("Failed to initialize tracer: %v", err)
	}
	defer shutdown(context.Background())

	// Initialize health checkers
	healthChecker := checker.NewHealthChecker(cfg, log)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(healthChecker, log)

	// Start background health checks
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go healthChecker.StartPeriodicChecks(ctx, 30*time.Second)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Health Service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// Health endpoints
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	})

	// Prometheus metrics
	app.Get("/metrics", func(c *fiber.Ctx) error {
		fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())(c.Context())
		return nil
	})

	// API routes
	api := app.Group("/api/v1")

	// Health check routes
	health := api.Group("/health")
	health.Get("/", healthHandler.GetAllHealth)
	health.Get("/summary", healthHandler.GetHealthSummary)
	health.Get("/:service", healthHandler.GetServiceHealth)
	health.Post("/check", healthHandler.TriggerHealthCheck)

	// Infrastructure health
	infra := api.Group("/infrastructure")
	infra.Get("/postgres", healthHandler.CheckPostgres)
	infra.Get("/kafka", healthHandler.CheckKafka)
	infra.Get("/redis", healthHandler.CheckRedis)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Infof("Health Service started on port %s", cfg.Port)

	<-quit
	log.Info("Shutting down server...")
	cancel()

	if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited")
}
