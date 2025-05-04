package cmd

import (
	"github.com/pmeier/redgiant/internal/serve"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the REST API",
	Run:   runFunc(serve.Run),
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
