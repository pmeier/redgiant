package live

import (
	"fmt"
	"os"

	"github.com/pmeier/redgiant/internal/redgiant"
	"github.com/pmeier/redgiant/internal/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LiveParams struct {
	SungrowHost     string
	SungrowPassword string
}

func Start(p LiveParams) error {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	rg := redgiant.NewRedGiant(p.SungrowHost, p.SungrowPassword)
	if err := rg.Connect(); err != nil {
		return err
	}
	defer rg.Close()

	d, err := utils.SummaryDevice(rg)
	if err != nil {
		return err
	}

	s, err := rg.Summary(d)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", s)
	return nil
}
