package handler

import (
	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/features/auth/service"
	"github.com/Testzyler/banking-api/server/exception"
	"github.com/Testzyler/banking-api/server/response"
	"github.com/gofiber/fiber/v2"
)

type authHandler struct {
	service service.AuthService
}

func NewAuthHandler(router fiber.Router, service service.AuthService) {
	handler := &authHandler{
		service: service,
	}

	auth := router.Group("/auth")
	auth.Post("/verify-pin", handler.VerifyPin)
	auth.Post("/refresh", handler.RefreshToken)
}

func (h *authHandler) VerifyPin(c *fiber.Ctx) error {
	var params entities.PinVerifyParams
	if err := c.BodyParser(&params); err != nil {
		return exception.ErrValidationFailed
	}

	if err := params.Validate(); err != nil {
		return err
	}

	tokenResponse, err := h.service.VerifyPin(params)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(&response.SuccessResponse{
		Code:    response.SuccessCodeOK,
		Message: "PIN verified successfully",
		Data:    tokenResponse,
	})
}

func (h *authHandler) RefreshToken(c *fiber.Ctx) error {
	var req entities.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Code:    response.ErrCodeBadRequest,
			Message: "Invalid request body",
			Details: err.Error(),
		})
	}

	tokenResponse, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		if errorResponse, ok := err.(*response.ErrorResponse); ok {
			return c.Status(errorResponse.HttpStatusCode).JSON(errorResponse)
		}

		return c.Status(fiber.StatusUnauthorized).JSON(response.ErrorResponse{
			Code:    response.ErrCodeUnauthorized,
			Message: "Invalid refresh token",
			Details: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Code:    response.SuccessCodeOK,
		Message: "Token refreshed successfully",
		Data:    tokenResponse,
	})
}
