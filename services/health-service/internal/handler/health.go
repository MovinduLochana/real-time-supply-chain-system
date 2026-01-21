package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"health-service/internal/checker"
)

type HealthHandler struct {
	checker *checker.HealthChecker
}

func NewHealthHandler(c *checker.HealthChecker) *HealthHandler {
	return &HealthHandler{checker: c}
}

func (h *HealthHandler) GetOverallHealth(c *fiber.Ctx) error {
	ctx := c.UserContext()
	_, span := otel.Tracer("health-handler").Start(ctx, "GetOverallHealth")
	defer span.End()

	health := h.checker.CheckAll(ctx)

	span.SetAttributes(attribute.String("health.status", string(health.Status)))

	statusCode := fiber.StatusOK
	if health.Status == checker.StatusUnhealthy {
		statusCode = fiber.StatusServiceUnavailable
	} else if health.Status == checker.StatusDegraded {
		statusCode = fiber.StatusMultiStatus
	}

	return c.Status(statusCode).JSON(health)
}

func (h *HealthHandler) GetInfrastructureHealth(c *fiber.Ctx) error {
	ctx := c.UserContext()
	_, span := otel.Tracer("health-handler").Start(ctx, "GetInfrastructureHealth")
	defer span.End()

	health := h.checker.CheckInfrastructure(ctx)

	span.SetAttributes(attribute.String("health.status", string(health.Status)))

	statusCode := fiber.StatusOK
	if health.Status == checker.StatusUnhealthy {
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(health)
}

func (h *HealthHandler) GetServicesHealth(c *fiber.Ctx) error {
	ctx := c.UserContext()
	_, span := otel.Tracer("health-handler").Start(ctx, "GetServicesHealth")
	defer span.End()

	health := h.checker.CheckServices(ctx)

	span.SetAttributes(attribute.String("health.status", string(health.Status)))

	statusCode := fiber.StatusOK
	if health.Status == checker.StatusUnhealthy {
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(health)
}

func (h *HealthHandler) GetLivenessProbe(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "alive",
	})
}

func (h *HealthHandler) GetReadinessProbe(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// Check infrastructure health for readiness
	health := h.checker.CheckInfrastructure(ctx)

	if health.Status == checker.StatusUnhealthy {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "not ready",
			"reason": "infrastructure unhealthy",
		})
	}

	return c.JSON(fiber.Map{
		"status": "ready",
	})
}

func (h *HealthHandler) GetLastCachedHealth(c *fiber.Ctx) error {
	lastCheck := h.checker.GetLastCheck()
	if lastCheck == nil {
		return c.Status(fiber.StatusNoContent).JSON(fiber.Map{
			"message": "no cached health check available",
		})
	}

	return c.JSON(lastCheck)
}

func (h *HealthHandler) RegisterRoutes(app *fiber.App) {
	health := app.Group("/health")
	health.Get("", h.GetOverallHealth)
	health.Get("/infrastructure", h.GetInfrastructureHealth)
	health.Get("/services", h.GetServicesHealth)
	health.Get("/live", h.GetLivenessProbe)
	health.Get("/ready", h.GetReadinessProbe)
	health.Get("/cached", h.GetLastCachedHealth)
}
