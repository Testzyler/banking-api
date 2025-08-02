package routes

import (
	authHandler "github.com/Testzyler/banking-api/app/features/auth/handler"
	authRepository "github.com/Testzyler/banking-api/app/features/auth/repository"
	authService "github.com/Testzyler/banking-api/app/features/auth/service"
	"github.com/Testzyler/banking-api/config"

	dashboardHandler "github.com/Testzyler/banking-api/app/features/dashboard/handler"
	dashboardRepository "github.com/Testzyler/banking-api/app/features/dashboard/repository"
	dashboardService "github.com/Testzyler/banking-api/app/features/dashboard/service"
	"github.com/Testzyler/banking-api/database"
	"github.com/gofiber/fiber/v2"
)

func InitHandlers(api fiber.Router, db database.DatabaseInterface) {

	// Register Dashboard handler
	dashboardHandler.NewDashboardHandler(
		api,
		dashboardService.NewDashboardService(
			dashboardRepository.NewDashboardRepository(db.GetDB()),
		),
	)

	// Register Auth handler
	authHandler.NewAuthHandler(
		api,
		authService.NewAuthService(
			authRepository.NewAuthRepository(db.GetDB()),
			authService.NewJwtService(config.GetConfig()),
			config.GetConfig(),
		),
	)
}
