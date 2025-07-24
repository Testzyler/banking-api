package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Testzyler/banking-api/config"
	"github.com/Testzyler/banking-api/server"
	"github.com/spf13/cobra"
)

var wg sync.WaitGroup
var httpServer *server.Server

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the API",
	Run: func(cmd *cobra.Command, args []string) {
		// Create root context with cancellation
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Load configuration
		config := config.NewConfig(configFile)

		// Start HTTP server
		initHttpServer(ctx, config)

		// Setup signal handling for graceful shutdown
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		// Wait for shutdown signal or context cancellation
		select {
		case sig := <-quit:
			log.Printf("Received shutdown signal: %v", sig)
		case <-ctx.Done():
			log.Println("Application context cancelled")
		}

		// Create shutdown context with timeout
		shutdownTimeout := time.Duration(config.Server.ShutdownTimeout) * time.Second
		if shutdownTimeout == 0 {
			shutdownTimeout = 30 * time.Second // Default 30 seconds
		}
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		// Channel to track shutdown completion
		shutdownComplete := make(chan struct{})

		// Perform shutdown in goroutine
		go func() {
			defer close(shutdownComplete)

			// Cancel the main context to signal all services to stop
			cancel()

			// Shutdown HTTP server
			if httpServer != nil {
				if err := httpServer.Shutdown(); err != nil {
					log.Printf("Error during HTTP server shutdown: %v", err)
				}
			}

			// Wait for all goroutines to finish
			wg.Wait()
		}()

		// Wait for either shutdown completion or timeout
		select {
		case <-shutdownComplete:
			log.Println("Graceful shutdown completed successfully")
		case <-shutdownCtx.Done():
			log.Printf("Shutdown timeout exceeded (%v), forcing exit", shutdownTimeout)
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
		log.Printf("Starting HTTP server on port %s", config.Server.Port)
		err := httpServer.Start()
		if err != nil {
			log.Printf("Error starting HTTP server: %v", err)
			return
		}
	}()
}
