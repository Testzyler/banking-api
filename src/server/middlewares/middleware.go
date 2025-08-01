package middlewares

import (
	"time"

	"github.com/Testzyler/banking-api/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")

		if requestID == "" {
			uuid, err := uuid.NewV7()
			if err != nil {
				return err
			}
			requestID = uuid.String()
		}

		// use in handlers
		c.Locals("requestID", requestID)

		// for clients
		c.Set("X-Request-ID", requestID)

		return c.Next()
	}
}

func GetRequestID(c *fiber.Ctx) string {
	if requestID, ok := c.Locals("requestID").(string); ok {
		return requestID
	}
	return ""
}

func GetStatus(c *fiber.Ctx) int {
	if status, ok := c.Locals("status").(int); ok {
		return status
	}
	return 0
}

func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		requestID := GetRequestID(c)
		requestLogger := logger.With("request_id", requestID)
		c.Locals("logger", requestLogger)

		err := c.Next()
		duration := time.Since(start)
		status := c.Context().Response.StatusCode()
		statusCode := GetStatus(c)

		if status >= 500 || statusCode >= 500 {
			requestLogger.Errorw("HTTP Response",
				"method", c.Method(),
				"path", c.Path(),
				"status", status,
				"duration", duration,
			)
		} else if status >= 400 || statusCode >= 400 {
			requestLogger.Warnw("HTTP Response",
				"method", c.Method(),
				"path", c.Path(),
				"status", status,
				"duration", duration,
			)
		} else {
			requestLogger.Infow("HTTP Response",
				"method", c.Method(),
				"path", c.Path(),
				"status", status,
				"duration", duration,
			)
		}

		return err
	}
}

func GetRequestLogger(c *fiber.Ctx) *zap.SugaredLogger {
	if requestLogger, ok := c.Locals("logger").(*zap.SugaredLogger); ok {
		return requestLogger
	}

	requestID := GetRequestID(c)
	return logger.With("request_id", requestID)
}
