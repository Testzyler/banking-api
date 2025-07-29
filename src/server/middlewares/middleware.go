package middlewares

import (
	"time"

	"github.com/Testzyler/banking-api/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestIDMiddleware generates a unique request ID for each incoming request
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if request ID is already present in headers
		requestID := c.Get("X-Request-ID")

		if requestID == "" {
			uuid, err := uuid.NewV7()
			if err != nil {
				return err
			}
			requestID = uuid.String()
		}

		// Set the request ID in the context for use in handlers
		c.Locals("requestID", requestID)

		// Set the request ID in response headers for clients
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

// LoggerMiddleware creates a middleware that logs HTTP requests with request ID
func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		requestID := GetRequestID(c)
		requestLogger := logger.With("request_id", requestID)
		c.Locals("logger", requestLogger)

		err := c.Next()
		duration := time.Since(start)
		status := c.Response().StatusCode()

		// Log the response with request ID
		if status >= 500 {
			requestLogger.Errorw("HTTP Response",
				"method", c.Method(),
				"path", c.Path(),
				"status", status,
				"duration_ms", duration.Milliseconds(),
			)
		} else if status >= 400 {
			requestLogger.Warnw("HTTP Response",
				"method", c.Method(),
				"path", c.Path(),
				"status", status,
				"duration_ms", duration.Milliseconds(),
			)
		} else {
			requestLogger.Infow("HTTP Response",
				"method", c.Method(),
				"path", c.Path(),
				"status", status,
				"duration_ms", duration.Milliseconds(),
			)
		}

		return err
	}
}

// GetRequestLogger returns the request-scoped logger with request ID
func GetRequestLogger(c *fiber.Ctx) *zap.SugaredLogger {
	if requestLogger, ok := c.Locals("logger").(*zap.SugaredLogger); ok {
		return requestLogger
	}

	requestID := GetRequestID(c)
	return logger.With("request_id", requestID)
}
