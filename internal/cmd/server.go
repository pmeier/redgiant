package cmd

import (
	"github.com/pmeier/redgiant/internal/server"
	"github.com/spf13/cobra"

	"github.com/rs/zerolog/log"
)

var serverViper = NewViper()

var serveCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts redgiant server",
	Run: func(cmd *cobra.Command, args []string) {
		p := server.ServerParams{}
		err := serverViper.Unmarshal(&p)
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		if err := server.Start(p); err != nil {
			log.Fatal().Err(err).Send()
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	defer serverViper.BindPFlags(serveCmd.Flags())

	serveCmd.Flags().String("sungrow-host", "", "Hostname of the Sungrow inverter (required)")
	serveCmd.Flags().String("sungrow-user", "user", "User of the Sungrow inverter")
	serveCmd.Flags().String("sungrow-password", "pw1111", "Password for the --sungrow-user account of the Sungrow inverter")

	serveCmd.Flags().String("host", "127.0.0.1", "Hostname to bind the redgiant server to")
	serveCmd.Flags().Uint("port", 8000, "Port to bind the redgiant server to")
}
