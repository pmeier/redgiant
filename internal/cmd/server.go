package cmd

import (
	"github.com/pmeier/redgiant/internal/server"
	"github.com/spf13/cobra"
)

var serverViper = NewViper()

var serveCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts redgiant server",
	Run: func(cmd *cobra.Command, args []string) {
		p := server.ServerParams{}
		err := serverViper.Unmarshal(&p)
		if err != nil {
			panic(err.Error())
		}

		if err := server.Start(p); err != nil {
			panic(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	defer serverViper.BindPFlags(serveCmd.Flags())

	serveCmd.Flags().String("sungrow-host", "", "Hostname of the Sungrow inverter (required)")
	serveCmd.Flags().String("sungrow-username", "user", "Username of the Sungrow inverter")
	serveCmd.Flags().String("sungrow-password", "pw1111", "Password for the --sungrow-user")

	serveCmd.Flags().String("host", "127.0.0.1", "Hostname to bind the redgiant server to")
	serveCmd.Flags().Uint("port", 8000, "Port to bind the redgiant server to")

	serveCmd.Flags().String("log-level", "info", "Log level")
}
