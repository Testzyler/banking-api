package dashboard

import (
	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/features/home/service"
	"github.com/Testzyler/banking-api/server/middlewares"
	"github.com/Testzyler/banking-api/server/response"
	"github.com/gofiber/fiber/v2"
)

type homeHandler struct {
	service service.HomeService
}

func NewHomeHandler(router fiber.Router, service service.HomeService) {
	handler := &homeHandler{
		service: service,
	}

	home := router.Group("/home")
	home.Get("/", middlewares.AuthMiddleware(), handler.GetHomeData)
}

func (h *homeHandler) GetHomeData(c *fiber.Ctx) error {
	userID := c.Locals("user").(*entities.Claims).UserID

	data, err := h.service.GetHomeData(userID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(&response.SuccessResponse{
		Code:    response.Success,
		Message: "Home screen data retrieved successfully",
		Data:    data,
	})
}
