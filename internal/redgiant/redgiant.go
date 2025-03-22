package redgiant

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type Redgiant struct {
	sg *Sungrow
}

func NewRedGiant(sungrowHost string, sungrowPassword string) *Redgiant {
	sg := NewSungrow(sungrowHost, sungrowPassword)
	return &Redgiant{sg: sg}
}

func (rg *Redgiant) Connect() error {
	log.WithField("host", rg.sg.Host).Info("connection established")
	return rg.sg.Connect()
}

func (rg *Redgiant) Close() {
	rg.sg.Close()
	log.WithField("host", rg.sg.Host).Info("connection closed")
}

func (rg *Redgiant) About() (About, error) {
	type data struct {
		RawDatapoints []RawDatapoint `json:"list"`
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

func (rg *Redgiant) Devices() ([]Device, error) {
	log.Trace("RedGiant.Devices()")
	type data struct {
		Devices []Device `json:"list"`
	}

	var d data
	if err := rg.sg.Send("devicelist",
		map[string]any{
			"is_check_token": "0",
			"type":           "0"},
		&d); err != nil {
		return nil, err
	}
	return d.Devices, nil
}

var serviceMap = map[int][]string{
	35: {"real", "real_battery"},
	44: {"real"},
}

func (rg *Redgiant) RawData(device Device) ([]RawDatapoint, error) {
	log.WithFields(log.Fields{"device": fmt.Sprintf("%+v", device)}).Trace("RedGiant.RawData()")

	services, ok := serviceMap[device.Type]
	if !ok {
		return nil, fmt.Errorf("unknown device type %d", device.Type)
	}

	datapoints := []RawDatapoint{}
	type data struct {
		Datapoints []RawDatapoint `json:"list"`
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

func (rg *Redgiant) Summary(device Device) (Summary, error) {
	log.WithFields(log.Fields{"device": fmt.Sprintf("%+v", device)}).Trace("RedGiant.RawData()")

	if device.Type != 35 {
		return Summary{}, errors.New("invalid device type for summary")
	}

	dps, err := rg.RawData(device)
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
