package handler

import (
	"log"
	"strconv"

	entities "github.com/Testzyler/banking-api/app/.entities"
	"github.com/Testzyler/banking-api/app/features/users/service"
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
	// Pagination parameters can be extracted from query params
	perPage := c.Query("per_page", "10")
	page := c.Query("page", "1")
	search := c.Query("search", "")

	perPageInt, err := strconv.Atoi(perPage)
	if err != nil || perPageInt <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid per_page parameter",
		})
	}
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid page parameter",
		})
	}

	params := entities.PaginationParams{
		Page:    pageInt,
		PerPage: perPageInt,
		Search:  search,
	}
	users, err := h.userService.GetAllUsers(params)
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}

	return c.JSON(users)
}

func (h *userHandler) GetUser(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	GetUserByIdParams := entities.GetUserByIdParams{
		UserID: userID,
	}
	if err := c.BodyParser(&GetUserByIdParams); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	user, err := h.userService.GetUserByID(GetUserByIdParams)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(user)
}
