package cmd

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Testzyler/banking-api/config"
	"github.com/Testzyler/banking-api/logger"
	"github.com/Testzyler/banking-api/server"
	"github.com/spf13/cobra"
)

var wg sync.WaitGroup
var httpServer *server.Server

var serveCmd = &cobra.Command{
	Use:   "serve_api",
	Short: "Serve the API",
	Run: func(cmd *cobra.Command, args []string) {
		// Create root context with cancellation
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Load configuration
		config := config.NewConfig(configFile)

		// Initialize logger
		loggerConfig := &logger.LoggerConfig{
			Level:       config.Logger.Level,
			Environment: config.Server.Environment,
			LogColor:    config.Logger.LogColor,
			LogJson:     config.Logger.LogJson,
		}
		if err := logger.InitLogger(loggerConfig); err != nil {
			panic("Failed to initialize logger: " + err.Error())
		}
		defer logger.SyncLogger()

		logger.Info("Starting Banking API Server", "port", config.Server.Port, "environment", config.Server.Environment)
		initHttpServer(ctx, config)

		// Setup signal handling for graceful shutdown
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		select {
		case sig := <-quit:
			logger.Info("Received shutdown signal", "signal", sig.String())
		case <-ctx.Done():
			logger.Info("Application context cancelled")
		}

		// Create shutdown context with timeout
		shutdownTimeout := time.Duration(config.Server.ShutdownTimeout) * time.Second
		if shutdownTimeout == 0 {
			shutdownTimeout = 30 * time.Second // Default 30 seconds
		}
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		shutdownComplete := make(chan struct{})
		go func() {
			defer close(shutdownComplete)
			cancel()
			if httpServer != nil {
				if err := httpServer.Shutdown(); err != nil {
					logger.Error("Error during HTTP server shutdown", "error", err)
				}
			}

			wg.Wait()
		}()

		logger.Info("Initiating graceful shutdown...")
		select {
		case <-shutdownComplete:
			logger.Info("Graceful shutdown completed successfully")
		case <-shutdownCtx.Done():
			logger.Error("Shutdown timeout exceeded, forcing exit", "timeout", shutdownTimeout)
			// Force exit if graceful shutdown takes too long
			os.Exit(1)
		}
	},
}

func init() {
	cmd.AddCommand(serveCmd)
}

func initHttpServer(ctx context.Context, config *config.Config) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Initialize server
		httpServer = server.NewServer(ctx, config)
		logger.Infof("HTTP server started on port %s", config.Server.Port)
		err := httpServer.Start()
		if err != nil {
			logger.Errorf("Error starting HTTP server %d : %v", config.Server.Port, err)
			return
		}
	}()
}
