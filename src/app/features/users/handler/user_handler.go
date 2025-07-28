package handler

import (
	"net/http"
	"strconv"

	entities "github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/features/users/service"
	"github.com/Testzyler/banking-api/server/exceptions"
	"github.com/Testzyler/banking-api/server/response"
	"github.com/gofiber/fiber/v2"
)

type userHandler struct {
	userService service.UserService
}

func NewUserHandler(router fiber.Router, userService service.UserService) {
	handler := &userHandler{
		userService: userService,
	}

	userGroup := router.Group("/users")

	userGroup.Get("/", handler.GetUsers)
	userGroup.Get("/:id", handler.GetUser)
}

func (h *userHandler) GetUsers(c *fiber.Ctx) error {
	// Parse pagination parameters
	var params entities.PaginationParams
	params.Page, _ = strconv.Atoi(c.Query("page", "1"))
	params.PerPage, _ = strconv.Atoi(c.Query("per_page", "10"))
	params.Search = c.Query("search", "")

	// if validationErr := exceptions.ValidateInput(&params); validationErr != nil {
	// 	return exceptions.SendBadRequestError(c, "Invalid pagination parameters", map[string]string{
	// 		"validation_error": validationErr.Message,
	// 	})
	// }

	users, err := h.userService.GetAllUsers(params)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(&response.ErrorResponse{
			BaseResponse: response.BaseResponse{
				Message: "Failed to fetch users",
			},
			Error: response.ErrorDetail{
				ErrorCode: exceptions.ErrCodeInternalServer,
				Details:   "An error occurred while retrieving users",
			},
		})
	}
	// Create pagination metadata (you'll need to update your service to return this info)
	meta := response.PaginationMeta{
		Page:    params.Page,
		PerPage: params.PerPage,
		// TODO: Get actual count from service
		Total:       len(users),
		TotalPages:  1,
		HasNext:     false,
		HasPrevious: params.Page > 1,
	}

	return c.Status(http.StatusOK).JSON(&response.PaginatedResponse{
		BaseResponse: response.BaseResponse{
			Message: "Users retrieved successfully",
			Data:    users,
		},
		Meta: meta,
	})
}

func (h *userHandler) GetUser(c *fiber.Ctx) error {
	var params entities.GetUserByIdParams
	params.UserID = c.Params("id")

	// if validationErr := exceptions.ValidateInput(&params); validationErr != nil {
	// 	return exceptions.SendBadRequestError(c, "Invalid parameters", map[string]string{
	// 		"validation_error": validationErr.Message,
	// 	})
	// }

	user, err := h.userService.GetUserByID(params)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(&response.ErrorResponse{
			BaseResponse: response.BaseResponse{
				Message: "User not found",
			},
			Error: response.ErrorDetail{
				ErrorCode: exceptions.ErrCodeNotFound,
				Details:   "The user with the specified ID does not exist",
			},
		})
	}

	return c.Status(http.StatusOK).JSON(&response.SuccessResponse{
		BaseResponse: response.BaseResponse{
			Message: "User retrieved successfully",
			Data:    user,
		},
	})

}
