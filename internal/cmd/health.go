package cmd

import (
	"github.com/pmeier/redgiant/internal/health"

	"github.com/spf13/cobra"
)

var healthViper = NewViper()

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check the health of the REST API",
	Run: func(cmd *cobra.Command, args []string) {
		p := health.HealthParams{}
		err := healthViper.Unmarshal(&p)
		if err != nil {
			panic(err.Error())
		}

		if err := health.Start(p); err != nil {
			panic(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)

	defer healthViper.BindPFlags(healthCmd.Flags())

	healthCmd.Flags().String("host", "127.0.0.1", "Hostname of the redgiant server")
	healthCmd.Flags().Uint("port", 8000, "Port of the redgiant server")
}
