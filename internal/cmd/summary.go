package cmd

import (
	"github.com/pmeier/redgiant/internal/summary"
	"github.com/spf13/cobra"
)

var summaryViper = NewViper()

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Get summary data",
	Run: func(cmd *cobra.Command, args []string) {
		p := summary.SummaryParams{}
		err := summaryViper.Unmarshal(&p)
		if err != nil {
			panic(err.Error())
		}

		if err := summary.Start(p); err != nil {
			panic(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(summaryCmd)
	defer summaryViper.BindPFlags(summaryCmd.Flags())

	summaryCmd.Flags().String("sungrow-host", "", "Hostname of the Sungrow inverter (required)")
	summaryCmd.Flags().String("sungrow-user", "user", "User of the Sungrow inverter")
	summaryCmd.Flags().String("sungrow-password", "pw1111", "Password for the --sungrow-user account of the Sungrow inverter")
	summaryCmd.Flags().Bool("quiet", true, "Don't output any logs")
	summaryCmd.Flags().Bool("json", false, "Output summary as JSON. Implies --quiet=true")
}
