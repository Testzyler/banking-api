package routes

import (
	authHandler "github.com/Testzyler/banking-api/app/features/auth/handler"
	authRepository "github.com/Testzyler/banking-api/app/features/auth/repository"
	authService "github.com/Testzyler/banking-api/app/features/auth/service"

	homeHandler "github.com/Testzyler/banking-api/app/features/home/handler"
	homeRepository "github.com/Testzyler/banking-api/app/features/home/repository"
	homeService "github.com/Testzyler/banking-api/app/features/home/service"
	"github.com/Testzyler/banking-api/config"

	"github.com/Testzyler/banking-api/database"
	"github.com/gofiber/fiber/v2"
)

func InitHandlers(api fiber.Router, db database.DatabaseInterface, redisDB *database.RedisDatabase) {
	// Register Home handler with AuthMiddleware protection
	homeHandler.NewHomeHandler(
		api,
		homeService.NewHomeService(
			homeRepository.NewHomeRepository(database.GetDatabase().GetDB()),
		),
	)

	// Register Auth handler
	authRepo := authRepository.NewAuthRepository(database.GetDatabase().GetDB(), database.GetCache())
	jwtService := authService.NewJwtService(config.GetConfig(), authRepo)
	authHandler.NewAuthHandler(
		api,
		authService.NewAuthService(
			authRepo,
			jwtService,
			config.GetConfig(),
		),
	)
}
