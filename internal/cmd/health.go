package cmd

import (
	"github.com/pmeier/redgiant/internal/health"

	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check the health of the REST API",
	Run:   runFunc(health.Run),
}

func init() {
	rootCmd.AddCommand(healthCmd)
}
