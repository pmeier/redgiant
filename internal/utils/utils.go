package utils

import (
	"errors"

	"github.com/pmeier/redgiant/internal/redgiant"
)

func Must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}

func SummaryDevice(rg *redgiant.Redgiant) (redgiant.Device, error) {
	devices, err := rg.Devices()
	if err != nil {
		panic(err.Error())
	}
	var device redgiant.Device
	for _, device = range devices {
		if device.Type == 35 {
			return device, nil
		}
	}

	return device, errors.New("no device suitable for summary found")
}
