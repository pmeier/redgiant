package cmd

import (
	"github.com/pmeier/redgiant/internal/live"
	"github.com/rs/zerolog/log"
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
			log.Fatal().Err(err).Send()
		}

		if err := live.Start(p); err != nil {
			log.Fatal().Err(err).Send()
		}
	},
}

func init() {
	rootCmd.AddCommand(liveCmd)
	defer liveViper.BindPFlags(liveCmd.Flags())

	liveCmd.Flags().String("sungrow-host", "", "Hostname of the Sungrow inverter (required)")
	liveCmd.Flags().String("sungrow-password", "pw1111", "Password for the user account of the Sungrow inverter")
}
