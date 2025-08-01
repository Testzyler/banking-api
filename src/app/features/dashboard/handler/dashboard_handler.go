package dashboard

import (
	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/features/dashboard/service"
	"github.com/Testzyler/banking-api/server/middlewares"
	"github.com/Testzyler/banking-api/server/response"
	"github.com/gofiber/fiber/v2"
)

type dashboardHandler struct {
	service service.DashboardService
}

func NewDashboardHandler(router fiber.Router, service service.DashboardService) {
	handler := &dashboardHandler{
		service: service,
	}

	dashboard := router.Group("/dashboard")
	dashboard.Get("/accounts", middlewares.AuthMiddleware(), handler.GetDashboardData)
}

func (h *dashboardHandler) GetDashboardData(c *fiber.Ctx) error {
	userID := c.Locals("user").(*entities.Claims).UserID

	// data, err := h.service.GetDashboardData(entities.DashboardParams{UserID: userID})
	// if err != nil {
	// 	return err
	// }
	
	// Optimized
	data, err := h.service.GetDashboardDataWithTrx(userID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(&response.SuccessResponse{
		Message: "Dashboard data retrieved successfully",
		Data:    data,
	})
}
