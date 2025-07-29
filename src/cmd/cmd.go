package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "banking-api",
	Short: "Start the Banking API server",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

// Execute root command
func Execute() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("Error executing root command: %v", err)
	}
}

var configFile string

func init() {
	cmd.PersistentFlags().StringVar(&configFile, "config", "config.yaml", "Config file (default is config.yaml)")
}
