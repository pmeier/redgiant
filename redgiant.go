package redgiant

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

type Redgiant struct {
	sg        *Sungrow
	log       zerolog.Logger
	localizer Localizer
	dm        map[int]Device
}

func NewRedgiant(sg *Sungrow, opts ...optFunc) *Redgiant {
	o := resolveOptions(append([]optFunc{WithLocalizer(NewSungrowLocalizer(sg.Host))}, opts...)...)
	return &Redgiant{sg: sg, log: o.logger, localizer: o.localizer}
}

func (rg *Redgiant) Connect() error {
	return rg.sg.Connect()
}

func (rg *Redgiant) Close() {
	rg.sg.Close()
}

func (rg *Redgiant) About() (About, error) {
	type data struct {
		RawDatapoints []Datapoint `json:"list"`
	}
	var d data
	if err := rg.sg.Get("/about/list", nil, &d); err != nil {
		return About{}, err
	}

	dps := map[string]string{}
	for _, dp := range d.RawDatapoints {
		dps[dp.I18nCode] = dp.Value
	}

	return About{
		SerialNumber:    dps["I18N_COMMON_DEVICE_SN"],
		Version:         dps["I18N_COMMON_VERSION"],
		SoftwareVersion: dps["I18N_COMMON_APPLI_SOFT_VERSION"],
		BuildVersion:    dps["I18N_COMMON_BUILD_SOFT_VERSION"],
	}, nil
}

func (rg *Redgiant) State() (State, error) {
	var s State
	if err := rg.sg.Send("state", nil, &s); err != nil {
		return State{}, err
	}
	return s, nil
}

func (rg *Redgiant) deviceMap() (map[int]Device, error) {
	rg.log.Trace().Msg("Redgiant.deviceMap()")

	if rg.dm == nil {
		type deviceList struct {
			Devices []Device `json:"list"`
		}

		var dl deviceList
		err := rg.sg.Send("devicelist",
			map[string]any{
				"is_check_token": "0",
				"type":           "0"},
			&dl)
		if err != nil {
			return nil, err
		}

		dm := map[int]Device{}
		for _, d := range dl.Devices {
			dm[d.ID] = d
		}
		rg.dm = dm
	}

	return rg.dm, nil
}

func (rg *Redgiant) Devices() ([]Device, error) {
	rg.log.Trace().Msg("Redgiant.Devices()")

	dm, err := rg.deviceMap()
	if err != nil {
		return nil, err
	}

	ds := make([]Device, 0, len(dm))
	for _, d := range dm {
		ds = append(ds, d)
	}

	return ds, nil
}

func (rg *Redgiant) getDevice(deviceID int) (Device, error) {
	dm, err := rg.deviceMap()
	if err != nil {
		return Device{}, err
	}

	d, ok := dm[deviceID]
	if !ok {
		msg := "unknown device"
		rg.log.Error().Int("deviceID", deviceID).Msg(msg)
		return Device{}, errors.New(msg)
	}

	return d, nil
}

var serviceMap = map[int][]string{
	35: {"real", "real_battery", "direct"},
	44: {"real"},
}

func (rg *Redgiant) LiveData(deviceID int, services ...string) ([]Datapoint, error) {
	rg.log.Trace().Int("deviceID", deviceID).Strs("services", services).Msg("Redgiant.LiveData()")

	device, err := rg.getDevice(deviceID)
	if err != nil {
		return nil, err
	}

	availableServices, ok := serviceMap[device.Type]
	if !ok {
		return nil, fmt.Errorf("unknown device type %d", device.Type)
	}
	if len(services) > 0 {
		for _, service := range services {
			if !slices.Contains(availableServices, service) {
				return nil, fmt.Errorf("unknown service %s for device type %d", service, device.Type)
			}
		}
	} else {
		services = availableServices
	}

	datapoints := []Datapoint{}
	type data struct {
		Datapoints []Datapoint `json:"list"`
	}
	var d data

	for _, service := range services {
		if err := rg.sg.Send(service, map[string]any{"dev_id": strconv.Itoa(device.ID), "time123456": time.Now().Unix()}, &d); err != nil {
			return nil, err
		}
		datapoints = append(datapoints, d.Datapoints...)
	}

	return datapoints, nil
}

func (rg *Redgiant) LocalizedLiveData(deviceID int, lang Language) ([]LocalizedDatapoint, error) {
	rg.log.Trace().Int("deviceID", deviceID).Stringer("lang", lang).Msg("Redgiant.LiveData()")

	if rg.localizer == nil {
		return nil, errors.New("no localizer available")
	}

	ld, err := rg.LiveData(deviceID)
	if err != nil {
		return nil, err
	}
	lld := make([]LocalizedDatapoint, 0, len(ld))
	for _, d := range ld {
		name, err := rg.localizer.Localize(d.I18nCode, lang)
		if err != nil {
			return nil, err
		}
		lld = append(lld, LocalizedDatapoint{Datapoint: d, Name: name})
	}
	return lld, nil
}
