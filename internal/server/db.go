package server

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pmeier/redgiant/internal/redgiant"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func newDB(host string, port uint, username string, password string, name string) *DB {
	dsn := compileDSN(host, port, username, password, name)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	db.AutoMigrate(&Quantity{}, &Data{})
	return &DB{DB: db}
}

func compileDSN(host string, port uint, username string, password string, name string) string {
	dsnKeyValues := map[string]string{
		"host":     host,
		"port":     strconv.Itoa(int(port)),
		"user":     username,
		"password": password,
		"dbname":   name,
		"sslmode":  "disable",
	}
	dsnPairs := make([]string, 0, len(dsnKeyValues))
	for key, value := range dsnKeyValues {
		dsnPairs = append(dsnPairs, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(dsnPairs, " ")
}

func (db *DB) Start(rg *redgiant.Redgiant, device redgiant.Device, interval time.Duration) {
	go func() {
		for timestamp := range time.NewTicker(interval).C {
			if err := db.store(rg, device, timestamp); err != nil {
				log.Error().Err(err).Send()
			}
		}
	}()
}

func (db *DB) store(rg *redgiant.Redgiant, device redgiant.Device, timestamp time.Time) error {
	s, err := rg.Summary(device)
	if err != nil {
		return err
	}

	log.Info().Msg("saving summary data")
	log.Trace().Time("timestamp", timestamp).Any("summary", s).Send()

	db.Create([]Data{
		{Timestamp: timestamp, QuantityID: 1, Value: s.GridPower},
		{Timestamp: timestamp, QuantityID: 2, Value: s.BatteryPower},
		{Timestamp: timestamp, QuantityID: 3, Value: s.PVPower},
		{Timestamp: timestamp, QuantityID: 4, Value: s.LoadPower},
		{Timestamp: timestamp, QuantityID: 5, Value: s.BatteryLevel},
	})

	return nil
}

type Quantity struct {
	ID    uint
	Name  string `gorm:"unique; not null"`
	Unit  string `gorm:"not null"`
	Datas []Data
}

type Data struct {
	ID         uint
	Timestamp  time.Time `gorm:"type:timestamptz(0); not null"`
	QuantityID uint      `gorm:"not null"`
	Value      float32   `gorm:"type:real; not null"`
}
