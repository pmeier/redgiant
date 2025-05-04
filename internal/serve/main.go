package serve

import (
	"os"
	"time"

	"github.com/pmeier/redgiant"
	"github.com/pmeier/redgiant/internal/config"

	"github.com/rs/zerolog"
)

func Run(c config.Config) error {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger().Level(c.LogLevel)

	sg := redgiant.NewSungrow(c.Sungrow.Host, c.Sungrow.Username, c.Sungrow.Password, redgiant.WithLogger(logger))
	rg := redgiant.NewRedgiant(sg, redgiant.WithLogger(logger))

	if err := rg.Connect(); err != nil {
		return err
	}
	defer rg.Close()

	s := newServer(rg, logger)
	if err := s.Start(c.Host, c.Port, 5*time.Second); err != nil {
		return err
	}

	select {}

	return nil
}
