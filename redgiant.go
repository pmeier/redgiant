package redgiant

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type RedgiantConfig struct {
	Logger zerolog.Logger
}

func DefaultRedgiantConfig() RedgiantConfig {
	return RedgiantConfig{Logger: log.Logger}
}

type Redgiant struct {
	sg        *Sungrow
	log       zerolog.Logger
	localizer Localizer
	dm        map[int]Device
}

func NewRedgiant(sg *Sungrow, config ...RedgiantConfig) *Redgiant {
	o := oneOptionalOrDefault(config, DefaultRedgiantConfig)
	return &Redgiant{sg: sg, log: o.Logger}
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

var serviceMap = map[int][]string{
	35: {"real", "real_battery"},
	44: {"real"},
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

func (rg *Redgiant) LiveData(deviceID int) ([]Datapoint, error) {
	rg.log.Trace().Int("deviceID", deviceID).Msg("Redgiant.LiveData()")

	device, err := rg.getDevice(deviceID)
	if err != nil {
		return nil, err
	}

	services, ok := serviceMap[device.Type]
	if !ok {
		return nil, fmt.Errorf("unknown device type %d", device.Type)
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

func (rg *Redgiant) Summary(deviceID int) (Summary, error) {
	rg.log.Trace().Int("deviceID", deviceID).Msg("Redgiant.Summary()")

	device, err := rg.getDevice(deviceID)
	if err != nil {
		return Summary{}, err
	}

	if device.Type != 35 {
		return Summary{}, errors.New("invalid device type for summary")
	}

	dps, err := rg.LiveData(device.ID)
	if err != nil {
		return Summary{}, err
	}
	vs := map[string]float32{}
	for _, dp := range dps {
		v, err := strconv.ParseFloat(dp.Value, 32)
		if err != nil {
			continue
		}
		vs[dp.I18nCode] = float32(v)
	}

	gridPower := (vs["I18N_CONFIG_KEY_4060"] - vs["I18N_COMMON_FEED_NETWORK_TOTAL_ACTIVE_POWER"]) * 1e3
	batteryPower := (vs["I18N_CONFIG_KEY_3921"] - vs["I18N_CONFIG_KEY_3907"]) * 1e3
	pvPower := vs["I18N_COMMON_TOTAL_DCPOWER"] * 1e3
	loadPower := vs["I18N_COMMON_LOAD_TOTAL_ACTIVE_POWER"] * 1e3

	batteryLevel := vs["I18N_COMMON_BATTERY_SOC"] * 1e-2

	return Summary{GridPower: gridPower, BatteryPower: batteryPower, PVPower: pvPower, LoadPower: loadPower, BatteryLevel: batteryLevel}, nil
}
