package serve

import (
	"time"

	"github.com/pmeier/redgiant"
	"github.com/pmeier/redgiant/internal/config"

	"github.com/rs/zerolog"
)

func Run(c config.Config) error {

	logger := zerolog.New(c.Logging.Format.Writer()).With().Timestamp().Logger().Level(c.Logging.Level)

	logger.Info().Msg("Look at this!")

	sg := redgiant.NewSungrow(c.Sungrow.Host, c.Sungrow.Username, c.Sungrow.Password, redgiant.WithLogger(logger))
	rg := redgiant.NewRedgiant(sg, redgiant.WithLogger(logger))

	if err := rg.Connect(); err != nil {
		return err
	}
	defer rg.Close()

	s := newServer(rg, logger)
	if err := s.Start(c.Server.Host, c.Server.Port, 5*time.Second); err != nil {
		return err
	}

	select {}

	return nil
}
