package server

import (
	"os"
	"time"

	"github.com/pmeier/redgiant"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ServerParams struct {
	SungrowHost     string
	SungrowUsername string
	SungrowPassword string
	Host            string
	Port            uint
}

func Start(p ServerParams) error {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	sg := redgiant.NewSungrow(p.SungrowHost, p.SungrowUsername, p.SungrowPassword)
	rg := redgiant.NewRedgiant(sg)
	if err := rg.Connect(); err != nil {
		return err
	}
	defer rg.Close()

	s := newServer(p, rg)
	if err := s.Start(5 * time.Second); err != nil {
		return err
	}

	select {}

	return nil
}
