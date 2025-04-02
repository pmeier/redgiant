package server

import (
	"fmt"
	"os"
	"time"

	"github.com/pmeier/redgiant"

	"github.com/rs/zerolog"
)

type ServerParams struct {
	SungrowHost     string
	SungrowUsername string
	SungrowPassword string
	Host            string
	Port            uint
	// FIXME: make this zerolog.LogLevel
	LogLevel string
}

func Start(p ServerParams) error {
	l, err := zerolog.ParseLevel(p.LogLevel)
	if err != nil {
		return fmt.Errorf("unknown log level %s", p.LogLevel)
	}

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger().Level(l)

	sg := redgiant.NewSungrow(p.SungrowHost, p.SungrowUsername, p.SungrowPassword, redgiant.WithLogger(logger))
	rg := redgiant.NewRedgiant(sg, redgiant.WithLogger(logger))

	if err := rg.Connect(); err != nil {
		return err
	}
	defer rg.Close()

	s := newServer(p, rg, logger)
	if err := s.Start(5 * time.Second); err != nil {
		return err
	}

	select {}

	return nil
}
