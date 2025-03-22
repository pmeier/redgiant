package cmd

import (
	"time"

	"github.com/pmeier/redgiant/internal/server"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var serverViper = NewViper()

var serveCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts redgiant server",
	Run: func(cmd *cobra.Command, args []string) {
		p := server.ServerParams{}
		err := serverViper.Unmarshal(&p)
		if err != nil {
			log.Fatal(err.Error())
		}

		if err := server.Start(p); err != nil {
			log.Error(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	defer serverViper.BindPFlags(serveCmd.Flags())

	serveCmd.Flags().String("sungrow-host", "", "Hostname of the Sungrow inverter (required)")
	serveCmd.Flags().String("sungrow-password", "pw1111", "Password for the user account of the Sungrow inverter")

	serveCmd.Flags().String("host", "127.0.0.1", "Hostname to bind the redgiant server to")
	serveCmd.Flags().Uint("port", 8000, "Port to bind the redgiant server to")

	serveCmd.Flags().Bool("database", false, "Store summary data periodically in a database")
	serveCmd.Flags().Duration("store-interval", time.Minute, "Store data this often")
	serveCmd.Flags().String("database-host", "127.0.0.1", "Hostname of the database")
	serveCmd.Flags().Uint("database-port", 5432, "Port of the database")
	serveCmd.Flags().String("database-username", "postgres", "Username of the database")
	serveCmd.Flags().String("database-password", "postgres", "Password of the database")
	serveCmd.Flags().String("database-name", "postgres", "Name of the database")
}
