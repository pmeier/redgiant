package redgiant

import (
	"encoding/json"
	"fmt"
)

type intBool bool

func (ib *intBool) UnmarshalJSON(data []byte) error {
	s := string(data)
	switch s {
	case "0":
		*ib = false
	case "1":
		*ib = true
	default:
		return fmt.Errorf("cannot unmarshal %s into bool", s)
	}
	return nil
}

type About struct {
	SerialNumber    string `json:"serialNumber"`
	Version         string `json:"version"`
	SoftwareVersion string `json:"softwareVersion"`
	BuildVersion    string `json:"buildVersion"`
}

type State struct {
	TotalFaults         int  `json:"totalFaults"`
	TotalAlarms         int  `json:"totalAlarms"`
	WirelessConnection  bool `json:"wirelessConnection"`
	WifiConnection      bool `json:"wifiConnection"`
	Ethernet1Connection bool `json:"ethernet1Connection"`
	Ethernet2Connection bool `json:"ethernet2Connection"`
	CloudConnection     bool `json:"cloudConnection"`
}

func (s *State) UnmarshalJSON(data []byte) error {
	type sungrowState struct {
		TotalFaults         int     `json:"total_fault,string"`
		TotalAlarms         int     `json:"total_alarm,string"`
		WirelessConnection  intBool `json:"wireless_conn_sts,string"`
		WifiConnection      intBool `json:"wifi_conn_sts,string"`
		Ethernet1Connection intBool `json:"eth_conn_sts,string"`
		Ethernet2Connection intBool `json:"eth2_conn_sts,string"`
		CloudConnection     intBool `json:"cloud_conn_sts,string"`
	}
	var ss sungrowState
	if err := json.Unmarshal(data, &ss); err != nil {
		return err
	}
	s.TotalFaults = ss.TotalFaults
	s.TotalAlarms = ss.TotalAlarms
	s.WirelessConnection = bool(ss.WirelessConnection)
	s.WifiConnection = bool(ss.WifiConnection)
	s.Ethernet1Connection = bool(ss.Ethernet1Connection)
	s.Ethernet2Connection = bool(ss.Ethernet2Connection)
	s.CloudConnection = bool(ss.CloudConnection)
	return nil
}

type Device struct {
	ID              int    `json:"id"`
	Code            int    `json:"code"`
	Type            int    `json:"type"`
	Protocol        int    `json:"protocol"`
	SerialNumber    string `json:"serialNumber"`
	Name            string `json:"name"`
	Model           string `json:"model"`
	Special         string `json:"special"`
	InvType         int    `json:"invType"`
	PortName        string `json:"portName"`
	PhysicalAddress int    `json:"physicalAddress"`
	LogicalAddress  int    `json:"logicalAddress"`
	LinkStatus      int    `json:"linkStatus"`
	InitStatus      int    `json:"initStatus"`
}

func (d *Device) UnmarshalJSON(data []byte) error {
	type sungrowDevice struct {
		ID              int    `json:"dev_id"`
		Code            int    `json:"dev_code"`
		Type            int    `json:"dev_type"`
		Protocol        int    `json:"dev_protocol"`
		SerialNumber    string `json:"dev_sn"`
		Name            string `json:"dev_name"`
		Model           string `json:"dev_model"`
		Special         string `json:"dev_special"`
		InvType         int    `json:"inv_type"`
		PortName        string `json:"port_name"`
		PhysicalAddress int    `json:"phys_addr,string"`
		LogicalAddress  int    `json:"logc_addr,string"`
		LinkStatus      int    `json:"link_status"`
		InitStatus      int    `json:"init_status"`
	}
	var sd sungrowDevice
	if err := json.Unmarshal(data, &sd); err != nil {
		return err
	}
	*d = Device(sd)
	return nil
}

type RealMeasurement struct {
	I18NCode string `json:"i18nCode"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Unit     string `json:"unit"`
}

func (rm *RealMeasurement) UnmarshalJSON(data []byte) error {
	type sungrowRealMeasurement struct {
		I18nCode string `json:"data_name"`
		Value    string `json:"data_value"`
		Unit     string `json:"data_unit"`
	}
	var srm sungrowRealMeasurement
	if err := json.Unmarshal(data, &srm); err != nil {
		return err
	}
	rm.I18NCode = srm.I18nCode
	rm.Value = srm.Value
	rm.Unit = srm.Unit
	return nil
}

type DirectMeasurement struct {
	I18NCode    string  `json:"i18nCode"`
	Name        string  `json:"name"`
	Voltage     float32 `json:"voltage"`
	VoltageUnit string  `json:"voltageUnit"`
	Current     float32 `json:"current"`
	CurrentUnit string  `json:"currentUnit"`
}

func (dm *DirectMeasurement) UnmarshalJSON(data []byte) error {
	type sungrowDirectMeasurement struct {
		I18NCode    string  `json:"name"`
		Voltage     float32 `json:"voltage,string"`
		VoltageUnit string  `json:"voltage_unit"`
		Current     float32 `json:"current,string"`
		CurrentUnit string  `json:"current_unit"`
	}
	var sdm sungrowDirectMeasurement
	if err := json.Unmarshal(data, &sdm); err != nil {
		return err
	}
	dm.I18NCode = sdm.I18NCode
	dm.Voltage = sdm.Voltage
	dm.VoltageUnit = sdm.VoltageUnit
	dm.Current = sdm.Current
	dm.CurrentUnit = sdm.CurrentUnit
	return nil
}
