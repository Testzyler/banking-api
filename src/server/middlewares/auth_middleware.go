package middlewares

import (
	"strings"

	authService "github.com/Testzyler/banking-api/app/features/auth/service"
	"github.com/Testzyler/banking-api/config"
	"github.com/Testzyler/banking-api/server/exception"
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return exception.ErrUnauthorized
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return exception.ErrUnauthorized
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			return exception.ErrUnauthorized
		}

		jwtService := authService.NewJwtService(config.GetConfig())

		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			return err
		}

		c.Locals("user", claims)
		return c.Next()
	}
}
