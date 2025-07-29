package routes

import (
	userHandler "github.com/Testzyler/banking-api/app/features/users/handler"
	userRepository "github.com/Testzyler/banking-api/app/features/users/repository"
	userService "github.com/Testzyler/banking-api/app/features/users/service"

	dashboardHandler "github.com/Testzyler/banking-api/app/features/dashboard/handler"
	dashboardRepository "github.com/Testzyler/banking-api/app/features/dashboard/repository"
	dashboardService "github.com/Testzyler/banking-api/app/features/dashboard/service"
	"github.com/Testzyler/banking-api/database"
	"github.com/gofiber/fiber/v2"
)

func InitHandlers(api fiber.Router, db database.DatabaseInterface) {
	// Register User handler
	userHandler.NewUserHandler(
		api,
		userService.NewUserService(
			userRepository.NewUserRepository(db.GetDB()),
		),
	)

	// Register Dashboard handler
	dashboardHandler.NewDashboardHandler(
		api,
		dashboardService.NewDashboardService(
			dashboardRepository.NewDashboardRepository(db.GetDB()),
		),
	)
}
