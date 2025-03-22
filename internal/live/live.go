package live

import (
	"fmt"

	"github.com/pmeier/redgiant/internal/redgiant"
	"github.com/pmeier/redgiant/internal/utils"

	log "github.com/sirupsen/logrus"
)

type LiveParams struct {
	SungrowHost     string
	SungrowPassword string
}

func Start(p LiveParams) error {
	log.SetLevel(log.InfoLevel)

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
