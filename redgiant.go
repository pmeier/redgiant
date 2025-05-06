package redgiant

import (
	"errors"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type deviceInfo struct {
	ID   int
	Type int
}

type Redgiant struct {
	sg            *Sungrow
	log           zerolog.Logger
	localizer     Localizer
	deviceInfoMap map[int]deviceInfo
}

func NewRedgiant(sg *Sungrow, opts ...OptFunc) *Redgiant {
	o := ResolveOptions(append([]OptFunc{
		WithLogger(log.Logger),
		WithLocalizer(NewSungrowLocalizer(sg.Host)),
	}, opts...)...)
	return &Redgiant{sg: sg, log: o.Logger, localizer: o.Localizer}
}

func (rg *Redgiant) Connect() error {
	return rg.sg.Connect()
}

func (rg *Redgiant) Close() {
	rg.sg.Close()
}

func (rg *Redgiant) About() (About, error) {
	type Data struct {
		Measurements []sungrowRealMeasurement `json:"list"`
	}
	var d Data
	if err := rg.sg.Get("/about/list", nil, &d); err != nil {
		return About{}, err
	}

	ms := map[string]string{}
	for _, m := range d.Measurements {
		ms[m.DataName] = m.DataValue
	}

	return About{
		SerialNumber:    ms["I18N_COMMON_DEVICE_SN"],
		Version:         ms["I18N_COMMON_VERSION"],
		SoftwareVersion: ms["I18N_COMMON_APPLI_SOFT_VERSION"],
		BuildVersion:    ms["I18N_COMMON_BUILD_SOFT_VERSION"],
	}, nil
}

func (rg *Redgiant) State() (State, error) {
	var s sungrowState
	if err := rg.sg.Send("state", nil, &s); err != nil {
		return State{}, err
	}
	return s.ToRedgiant(), nil
}

func (rg *Redgiant) Devices() ([]Device, error) {
	rg.log.Trace().Msg("Redgiant.Devices()")

	type Data struct {
		Devices []sungrowDevice `json:"list"`
	}
	var d Data
	err := rg.sg.Send("devicelist",
		map[string]any{
			"is_check_token": "0",
			"type":           "0"},
		&d)
	if err != nil {
		return nil, err
	}

	ds := make([]Device, 0, len(d.Devices))
	for _, sd := range d.Devices {
		ds = append(ds, sd.ToRedgiant())
	}

	return ds, nil
}

func (rg *Redgiant) getDeviceInfo(deviceID int) (deviceInfo, error) {
	rg.log.Trace().Msg("Redgiant.getDeviceInfo()")

	if rg.deviceInfoMap == nil {
		devices, err := rg.Devices()
		if err != nil {
			return deviceInfo{}, err
		}

		deviceInfoMap := make(map[int]deviceInfo, len(devices))
		for _, d := range devices {
			deviceInfoMap[d.ID] = deviceInfo{ID: d.ID, Type: d.Type}
		}
		rg.deviceInfoMap = deviceInfoMap
	}

	i, ok := rg.deviceInfoMap[deviceID]
	if !ok {
		msg := "unknown device"
		rg.log.Error().Int("deviceID", deviceID).Msg(msg)
		return deviceInfo{}, errors.New(msg)
	}

	return i, nil
}

var availableRealDataServices = map[int][]string{
	35: {"real", "real_battery"},
	44: {"real"},
}

func (rg *Redgiant) RealData(deviceID int, lang Language, services ...string) ([]RealMeasurement, error) {
	rg.log.Trace().Int("deviceID", deviceID).Stringer("lang", lang).Strs("services", services).Msg("Redgiant.RealData()")

	info, err := rg.getDeviceInfo(deviceID)
	if err != nil {
		return nil, err
	}

	var strict bool
	if len(services) == 0 {
		var ok bool
		services, ok = availableRealDataServices[info.Type]
		if !ok {
			return nil, errors.New("unknown device type")
		}
		strict = false
	} else {
		strict = true
	}

	type Data struct {
		Measurements []sungrowRealMeasurement `json:"list"`
	}
	var d Data
	ms := []RealMeasurement{}
	for _, service := range services {
		if err := rg.sg.Send(service, map[string]any{"dev_id": strconv.Itoa(info.ID), "time123456": time.Now().Unix()}, &d); err != nil {
			if strict {
				return nil, err
			} else {
				continue
			}

		}
		for _, sm := range d.Measurements {
			m := sm.ToRedgiant()
			if name, err := rg.localizer.Localize(m.I18NCode, lang); err == nil {
				m.Name = name
			} else {
				m.Name = m.I18NCode
			}

			value, err := rg.localizer.Localize(m.Value, lang)
			if err == nil {
				m.Value = value
			}

			ms = append(ms, m)
		}
	}

	return ms, nil
}

var availableDirectDataServices = map[int][]string{
	35: {"direct"},
}

func (rg *Redgiant) DirectData(deviceID int, lang Language, services ...string) ([]DirectMeasurement, error) {
	rg.log.Trace().Int("deviceID", deviceID).Stringer("lang", lang).Strs("services", services).Msg("Redgiant.DirectData()")

	info, err := rg.getDeviceInfo(deviceID)
	if err != nil {
		return nil, err
	}

	var strict bool
	if len(services) == 0 {
		var ok bool
		services, ok = availableDirectDataServices[info.Type]
		if !ok {
			return nil, errors.New("unknown device type")
		}
		strict = false
	} else {
		strict = true
	}

	type Data struct {
		Measurements []sungrowDirectMeasurement `json:"list"`
	}
	var d Data
	ms := []DirectMeasurement{}
	for _, service := range services {
		if err := rg.sg.Send(service, map[string]any{"dev_id": strconv.Itoa(info.ID), "time123456": time.Now().Unix()}, &d); err != nil {
			if strict {
				return nil, err
			} else {
				continue
			}
		}
		for _, sm := range d.Measurements {
			m := sm.ToRedgiant()
			name, err := rg.localizer.Localize(m.I18NCode, lang)
			if err == nil {
				m.Name = name
			}

			ms = append(ms, m)
		}
	}

	return ms, nil
}
