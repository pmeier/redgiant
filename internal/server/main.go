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

	logger := zerolog.Logger{}.Output(zerolog.ConsoleWriter{Out: os.Stdout}).Level(l)

	sgc := redgiant.DefaultSungrowConfig()
	sgc.Logger = logger
	sg := redgiant.NewSungrow(p.SungrowHost, p.SungrowUsername, p.SungrowPassword, sgc)

	rgc := redgiant.DefaultRedgiantConfig()
	sgc.Logger = logger
	rg := redgiant.NewRedgiant(sg, rgc)

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
