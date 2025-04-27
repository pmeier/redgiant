package redgiant

import (
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

type sungrowState struct {
	TotalFault      int     `json:"total_fault,string"`
	TotalAlarm      int     `json:"total_alarm,string"`
	WirelessConnSts intBool `json:"wireless_conn_sts,string"`
	WifiConnSts     intBool `json:"wifi_conn_sts,string"`
	EthConnSts      intBool `json:"eth_conn_sts,string"`
	Eth2ConnSts     intBool `json:"eth2_conn_sts,string"`
	CloudConnSts    intBool `json:"cloud_conn_sts,string"`
}

func (ss *sungrowState) ToRedgiant() State {
	return State{
		TotalFaults:         ss.TotalFault,
		TotalAlarms:         ss.TotalAlarm,
		WirelessConnection:  bool(ss.WirelessConnSts),
		WifiConnection:      bool(ss.WifiConnSts),
		Ethernet1Connection: bool(ss.EthConnSts),
		Ethernet2Connection: bool(ss.Eth2ConnSts),
		CloudConnection:     bool(ss.CloudConnSts),
	}
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

type sungrowDevice struct {
	DevID       int    `json:"dev_id"`
	DevCode     int    `json:"dev_code"`
	DevType     int    `json:"dev_type"`
	DevProtocol int    `json:"dev_protocol"`
	DevSN       string `json:"dev_sn"`
	DevName     string `json:"dev_name"`
	DevModel    string `json:"dev_model"`
	DevSpecial  string `json:"dev_special"`
	InvType     int    `json:"inv_type"`
	PortName    string `json:"port_name"`
	PhysAddress int    `json:"phys_addr,string"`
	LogcAddress int    `json:"logc_addr,string"`
	LinkStatus  int    `json:"link_status"`
	InitStatus  int    `json:"init_status"`
}

func (sd *sungrowDevice) ToRedgiant() Device {
	return Device{
		ID:              sd.DevID,
		Code:            sd.DevCode,
		Type:            sd.DevType,
		Protocol:        sd.DevProtocol,
		SerialNumber:    sd.DevSN,
		Name:            sd.DevName,
		Model:           sd.DevModel,
		Special:         sd.DevSpecial,
		InvType:         sd.InvType,
		PortName:        sd.PortName,
		PhysicalAddress: sd.PhysAddress,
		LogicalAddress:  sd.LogcAddress,
		LinkStatus:      sd.LinkStatus,
		InitStatus:      sd.InitStatus,
	}
}

type RealMeasurement struct {
	I18NCode string `json:"i18nCode"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Unit     string `json:"unit"`
}

type sungrowRealMeasurement struct {
	DataName  string `json:"data_name"`
	DataValue string `json:"data_value"`
	DataUnit  string `json:"data_unit"`
}

func (srm *sungrowRealMeasurement) ToRedgiant() RealMeasurement {
	return RealMeasurement{
		I18NCode: srm.DataName,
		Name:     "",
		Value:    srm.DataValue,
		Unit:     srm.DataUnit,
	}
}

type DirectMeasurement struct {
	I18NCode    string  `json:"i18nCode"`
	Name        string  `json:"name"`
	Voltage     float32 `json:"voltage"`
	VoltageUnit string  `json:"voltageUnit"`
	Current     float32 `json:"current"`
	CurrentUnit string  `json:"currentUnit"`
}

type sungrowDirectMeasurement struct {
	Name        string  `json:"name"`
	Voltage     float32 `json:"voltage,string"`
	VoltageUnit string  `json:"voltage_unit"`
	Current     float32 `json:"current,string"`
	CurrentUnit string  `json:"current_unit"`
}

func (sdm *sungrowDirectMeasurement) ToRedgiant() DirectMeasurement {
	return DirectMeasurement{
		I18NCode:    sdm.Name,
		Name:        "",
		Voltage:     sdm.Voltage,
		VoltageUnit: sdm.VoltageUnit,
		Current:     sdm.Current,
		CurrentUnit: sdm.CurrentUnit,
	}
}
