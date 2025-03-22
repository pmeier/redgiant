package cmd

import (
	"github.com/pmeier/redgiant/internal/health"

	"github.com/spf13/cobra"
)

var hp = health.HealthParams{}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Health check",
	Run: func(cmd *cobra.Command, args []string) {
		health.Start(hp)
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)

	healthCmd.Flags().StringVar(&hp.RedgiantHost, "redgiant-host", "127.0.0.1", "Hostname of the redgiant server")
	healthCmd.Flags().UintVar(&hp.RedgiantPort, "redgiant-port", 8000, "Port of the redgiant server")
}
