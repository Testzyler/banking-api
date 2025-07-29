package dashboard

import (
	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/features/dashboard/service"
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
	dashboard.Get("/accounts", handler.GetDashboardData)
}

func (h *dashboardHandler) GetDashboardData(c *fiber.Ctx) error {
	var params entities.DashboardParams
	if err := c.BodyParser(&params); err != nil {
		return err
	}
	if err := params.Validate(); err != nil {
		return err
	}

	data, err := h.service.GetDashboardData(params)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(&response.SuccessResponse{
		Message: "Dashboard data retrieved successfully",
		Data:    data,
	})
}
