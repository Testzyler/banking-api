package middlewares

import (
	"strings"

	authRepository "github.com/Testzyler/banking-api/app/features/auth/repository"
	authService "github.com/Testzyler/banking-api/app/features/auth/service"
	"github.com/Testzyler/banking-api/config"
	"github.com/Testzyler/banking-api/database"
	"github.com/Testzyler/banking-api/logger"
	"github.com/Testzyler/banking-api/server/exception"
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			c.Locals("status", fiber.StatusUnauthorized)
			return exception.ErrUnauthorized
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.Locals("status", fiber.StatusUnauthorized)
			return exception.ErrUnauthorized
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.Locals("status", fiber.StatusUnauthorized)
			return exception.ErrUnauthorized
		}

		authRepo := authRepository.NewAuthRepository(database.GetDatabase().GetDB(), database.GetCache())
		jwtService := authService.NewJwtService(config.GetConfig(), authRepo)
		validationResult, err := jwtService.ValidateTokenWithBanCheck(token)
		if err != nil {
			logger.Warnf("Token validation failed: %v", err)
			c.Locals("status", fiber.StatusUnauthorized)
			return err
		}

		if !validationResult.Valid {
			logger.Infof("Blocked request with invalid token, reason: %s", validationResult.Reason)
			c.Locals("status", fiber.StatusUnauthorized)
			return exception.ErrUnauthorized
		}

		c.Locals("user", validationResult.Claims)
		return c.Next()
	}
}
