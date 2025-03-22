package cmd

import (
	"os"

	"github.com/pmeier/redgiant/internal/live"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var liveViper = NewViper()

var liveCmd = &cobra.Command{
	Use:   "live",
	Short: "Get live data",
	Run: func(cmd *cobra.Command, args []string) {
		p := live.LiveParams{}
		err := liveViper.Unmarshal(&p)
		if err != nil {
			log.Fatal(err.Error())
			os.Exit(1)
		}

		if err := live.Start(p); err != nil {
			log.Fatal(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(liveCmd)
	defer liveViper.BindPFlags(liveCmd.Flags())

	liveCmd.Flags().String("sungrow-host", "", "Hostname of the Sungrow inverter (required)")
	liveCmd.Flags().String("sungrow-password", "pw1111", "Password for the user account of the Sungrow inverter")
}
