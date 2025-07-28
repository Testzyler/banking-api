package server

import (
	"context"
	"log"
	"time"

	"github.com/Testzyler/banking-api/config"
	"github.com/Testzyler/banking-api/database"
	"github.com/Testzyler/banking-api/server/exceptions"
	"github.com/Testzyler/banking-api/server/response"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	userHandler "github.com/Testzyler/banking-api/app/features/users/handler"
	userRepository "github.com/Testzyler/banking-api/app/features/users/repository"
	userService "github.com/Testzyler/banking-api/app/features/users/service"
)

type Server struct {
	App            *fiber.App
	Config         *config.Config
	DB             database.DatabaseInterface
	isShuttingDown bool
}

func NewServer(ctx context.Context, config *config.Config) *Server {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
		ReadTimeout:           config.Server.ReadTimeout,
		WriteTimeout:          config.Server.WriteTimeout,
		IdleTimeout:           config.Server.IdleTimeout,
		Concurrency:           config.Server.MaxConnections * 256 * 1024,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("Error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(&response.ErrorResponse{
				BaseResponse: response.BaseResponse{
					Message: "Internal Server Error",
				},
				Error: response.ErrorDetail{
					ErrorCode: exceptions.ErrCodeInternalServer,
					Details:   "An unexpected error occurred",
				},
			})
		},
	})

	// Initialize database
	db, err := database.NewDatabase(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	server := &Server{
		App:            app,
		Config:         config,
		DB:             db,
		isShuttingDown: false,
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

// Middleware
func (s *Server) setupMiddleware() {
	// Global error handling middleware
	// s.App.Use(pkg.GlobalErrorMiddleware())

	// Recovery middleware
	s.App.Use(recover.New())

	// Logger middleware
	s.App.Use(logger.New())

	// CORS middleware
	s.App.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "*",
	}))
}

func (s *Server) setupRoutes() {
	// API routes
	api := s.App.Group("/api/v1")

	// Health check
	api.Get("/healthz", func(c *fiber.Ctx) error {
		if s.isShuttingDown {
			return c.Status(fiber.StatusServiceUnavailable).JSON(response.ErrorResponse{
				Error: response.ErrorDetail{
					ErrorCode: exceptions.ErrCodeServiceUnavailable,
					Details:   "Service is unavailable",
				},
			})
		}

		healthData := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
		}

		return c.Status(fiber.StatusOK).JSON(&response.SuccessResponse{
			BaseResponse: response.BaseResponse{
				Data:    healthData,
				Message: "Health check successful",
			},
		})
	})

	// Register User handler
	userHandler.NewUserHandler(
		api,
		userService.NewUserService(
			userRepository.NewUserRepository(s.DB.GetDB()),
		),
	)

	// Setup 404 handler
	s.App.Use(response.NotFoundHandler)
}

func (s *Server) Start() error {
	if err := s.App.Listen(":" + s.Config.Server.Port); err != nil {
		return err
	}
	return nil
}

func (s *Server) Shutdown() error {
	var shutdownErrors []error
	s.isShuttingDown = true

	// Shut down the Fiber application
	if err := s.App.Shutdown(); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
		shutdownErrors = append(shutdownErrors, err)
	} else {
		log.Println("HTTP server shutdown successfully")
	}

	// Close database connections
	if s.DB != nil {
		if err := s.DB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
			shutdownErrors = append(shutdownErrors, err)
		} else {
			log.Println("Database connections closed successfully")
		}
	}

	// Return first error if any occurred
	if len(shutdownErrors) > 0 {
		return shutdownErrors[0]
	}
	return nil
}
