package server

import (
	"context"
	"time"

	"github.com/Testzyler/banking-api/config"
	"github.com/Testzyler/banking-api/database"
	"github.com/Testzyler/banking-api/logger"
	"github.com/Testzyler/banking-api/server/exception"
	"github.com/Testzyler/banking-api/server/middlewares"
	"github.com/Testzyler/banking-api/server/response"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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
		DisableStartupMessage: true,
		ReadTimeout:           config.Server.ReadTimeout,
		WriteTimeout:          config.Server.WriteTimeout,
		IdleTimeout:           config.Server.IdleTimeout,
		Concurrency:           config.Server.MaxConnections * 256 * 1024,
		ErrorHandler:          middlewares.ErrorHandler(), // Use the new error handler middleware
	})

	// Initialize database
	db, err := database.NewDatabase(config)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
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
	// Request ID middleware
	s.App.Use(middlewares.RequestIDMiddleware())

	// Recovery middleware
	s.App.Use(recover.New())

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
	s.App.Get("/healthz", func(c *fiber.Ctx) error {
		if s.isShuttingDown {
			return exception.ErrServiceUnavailable
		}

		healthData := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now(),
		}

		return c.JSON(&response.SuccessResponse{
			Message: "Service is healthy",
			Code:    response.SuccessCodeOK,
			Data:    healthData,
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
	s.App.Use(middlewares.NotFoundHandler)
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
		logger.Error("Error shutting down HTTP server", "error", err)
		shutdownErrors = append(shutdownErrors, err)
	} else {
		logger.Info("HTTP server shutdown successfully")
	}

	// Close database connections
	if s.DB != nil {
		if err := s.DB.Close(); err != nil {
			logger.Error("Error closing database", "error", err)
			shutdownErrors = append(shutdownErrors, err)
		} else {
			logger.Info("Database connections closed successfully")
		}
	}

	// Return first error if any occurred
	if len(shutdownErrors) > 0 {
		return shutdownErrors[0]
	}
	return nil
}
