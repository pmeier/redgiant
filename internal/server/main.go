package server

import (
	"time"

	"github.com/pmeier/redgiant/internal/redgiant"
	"github.com/pmeier/redgiant/internal/utils"

	log "github.com/sirupsen/logrus"
)

type ServerParams struct {
	SungrowHost      string
	SungrowPassword  string
	Host             string
	Port             uint
	Database         bool
	StoreInterval    time.Duration
	DatabaseHost     string
	DatabasePort     uint
	DatabaseUsername string
	DatabasePassword string
	DatabaseName     string
}

func Start(p ServerParams) error {
	log.SetLevel(log.DebugLevel)

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

	if p.Database {
		db := newDB(p.DatabaseHost, p.DatabasePort, p.DatabaseUsername, p.DatabasePassword, p.DatabaseName)
		db.Start(rg, device, p.StoreInterval)
	}

	select {}

	return nil
}
