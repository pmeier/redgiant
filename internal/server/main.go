package server

import (
	"os"
	"time"

	"github.com/pmeier/redgiant/internal/redgiant"
	"github.com/pmeier/redgiant/internal/utils"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ServerParams struct {
	SungrowHost     string
	SungrowPassword string
	Host            string
	Port            uint
}

func Start(p ServerParams) error {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	rg := redgiant.NewRedGiant(p.SungrowHost, p.SungrowPassword)
	if err := rg.Connect(); err != nil {
		return err
	}
	defer rg.Close()

	device := utils.Must(utils.SummaryDevice(rg))

	s := newServer(p, rg, device)
	if err := s.Start(5 * time.Second); err != nil {
		return err
	}

	select {}

	return nil
}
