package summary

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/pmeier/redgiant/internal/redgiant"
	"github.com/rs/zerolog"
)

type SummaryParams struct {
	SungrowHost     string
	SungrowPassword string
	Quiet           bool
	JSON            bool
}

func Start(p SummaryParams) error {
	if p.JSON || p.Quiet {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}

	rg := redgiant.NewRedGiant(p.SungrowHost, p.SungrowPassword)
	if err := rg.Connect(); err != nil {
		return err
	}
	defer rg.Close()

	// FIXME: don't hardcode this
	s, err := rg.Summary(0)
	if err != nil {
		return err
	}

	if p.JSON {
		b, err := json.Marshal(s)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	} else {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "Grid Power\t%+4.1f kW\n", s.GridPower*1e-3)
		fmt.Fprintf(w, "Battery Power\t%+4.1f kW\n", s.BatteryPower*1e-3)
		fmt.Fprintf(w, "PV Power\t%+4.1f kW\n", s.PVPower*1e-3)
		fmt.Fprintf(w, "Load Power\t%+4.1f kW\n", s.LoadPower*1e-3)
		fmt.Fprintf(w, "Battery Level\t%4.1f %%\n", s.BatteryLevel*1e2)
		w.Flush()
	}

	return nil
}
