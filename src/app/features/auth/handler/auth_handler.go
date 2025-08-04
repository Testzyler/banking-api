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
	auth.Post("/tokens", handler.BanAllUserTokens)
	auth.Get("/tokens", handler.ListAllTokens)
}

func (h *authHandler) ListAllTokens(c *fiber.Ctx) error {
	tokens, err := h.service.ListTokens(c.Context())
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(&response.SuccessResponse{
		Code:    response.Success,
		Message: "Tokens retrieved successfully",
		Data:    tokens,
	})
}

func (h *authHandler) VerifyPin(c *fiber.Ctx) error {
	var params entities.PinVerifyParams
	if err := c.BodyParser(&params); err != nil {
		return exception.ErrValidationFailed
	}

	if err := params.Validate(); err != nil {
		return err
	}

	tokenResponse, err := h.service.VerifyPin(c.Context(), params)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(&response.SuccessResponse{
		Code:    response.Success,
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
		Code:    response.Success,
		Message: "Token refreshed successfully",
		Data:    tokenResponse,
	})
}

func (h *authHandler) BanAllUserTokens(c *fiber.Ctx) error {
	var req entities.BanTokensRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Code:    response.ErrCodeBadRequest,
			Message: "User ID is required",
		})
	}

	if err := h.service.BanToken(c.Context(), req.UserID); err != nil {
		if errorResponse, ok := err.(*response.ErrorResponse); ok {
			return c.Status(errorResponse.HttpStatusCode).JSON(errorResponse)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Code:    response.ErrCodeInternalServer,
			Message: "Failed to ban token",
			Details: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Code:    response.Success,
		Message: "Token banned successfully",
	})
}
