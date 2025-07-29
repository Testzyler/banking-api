package cmd

import (
	"fmt"

	"github.com/Testzyler/banking-api/config"
	"github.com/Testzyler/banking-api/database"
	"github.com/spf13/cobra"
)

// MigrateCmd is the command to run database migrations
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  "This command runs the database migrations to ensure the database schema is up to date.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		config := config.NewConfig(configFile)

		// Initialize database connection
		db, err := database.NewDatabase(config)
		if err != nil {
			return fmt.Errorf("failed to get database connection: %w", err)
		}
		defer db.Close()

		// Run migrations
		fmt.Println("Starting database migrations...")
		if err := db.RunMigrations(); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		fmt.Println("Database migrations completed successfully.")
		return nil
	},
}

func init() {
	cmd.AddCommand(migrateCmd)
}
