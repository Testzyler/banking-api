package handler

import (
	"net/http"
	"strconv"

	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/features/users/service"
	"github.com/Testzyler/banking-api/server/middlewares"
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
	// Get request logger with request ID
	requestLogger := middlewares.GetRequestLogger(c)
	requestLogger.Infof("Getting users list %v", "Fetching users with pagination")

	var params entities.PaginationParams
	params.Page, _ = strconv.Atoi(c.Query("page", "1"))
	params.PerPage, _ = strconv.Atoi(c.Query("perPage", "10"))
	params.Search = c.Query("search", "")

	if err := params.Validate(); err != nil {
		return err
	}

	users, err := h.userService.GetAllUsers(params)
	if err != nil {
		return &response.ErrorResponse{
			HttpStatusCode: 500,
			Code:           response.ErrCodeInternalServer,
			Message:        "Failed to retrieve user",
			Details:        err.Error(),
		}
	}

	// Create pagination metadata (you'll need to update your service to return this info)
	meta := entities.PaginationMeta{
		Page:        params.Page,
		PerPage:     params.PerPage,
		Total:       len(users),
		TotalPages:  1,
		HasNext:     false,
		HasPrevious: params.Page > 1,
	}

	return c.Status(http.StatusOK).JSON(&entities.PaginatedResponse{
		SuccessResponse: response.SuccessResponse{
			Message: "Users retrieved successfully",
			Data:    users,
		},
		Meta: meta,
	})
}

func (h *userHandler) GetUser(c *fiber.Ctx) error {
	var params entities.GetUserByIdParams
	params.UserID = c.Params("id")

	user, err := h.userService.GetUserByID(params)
	if err != nil {
		return &response.ErrorResponse{
			HttpStatusCode: 404,
			Code:           response.ErrCodeNotFound,
			Details:        err.Error(),
		}
	}

	return c.Status(http.StatusOK).JSON(&response.SuccessResponse{
		Message: "User retrieved successfully",
		Data:    user,
	})

}
