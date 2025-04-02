package redgiant

import (
	"encoding/json"
	"fmt"
)

type IntBool bool

func (ib *IntBool) UnmarshalJSON(data []byte) error {
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
		WirelessConnection  IntBool `json:"wireless_conn_sts,string"`
		WifiConnection      IntBool `json:"wifi_conn_sts,string"`
		Ethernet1Connection IntBool `json:"eth_conn_sts,string"`
		Ethernet2Connection IntBool `json:"eth2_conn_sts,string"`
		CloudConnection     IntBool `json:"cloud_conn_sts,string"`
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

// FIXME: add available services to the device
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
	d.ID = sd.ID
	d.Code = sd.Code
	d.Type = sd.Type
	d.Protocol = sd.Protocol
	d.SerialNumber = sd.SerialNumber
	d.Name = sd.Name
	d.Model = sd.Model
	d.Special = sd.Special
	d.InvType = sd.InvType
	d.PortName = sd.PortName
	d.PhysicalAddress = sd.PhysicalAddress
	d.LogicalAddress = sd.LogicalAddress
	d.LinkStatus = sd.LinkStatus
	d.InitStatus = sd.InitStatus
	return nil
}

type Datapoint struct {
	I18nCode string `json:"i18nCode"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Unit     string `json:"unit"`
}

func (d *Datapoint) UnmarshalJSON(data []byte) error {
	type sungrowDatapoint struct {
		I18nCode string `json:"data_name"`
		Value    string `json:"data_value"`
		Unit     string `json:"data_unit"`
	}
	var sd sungrowDatapoint
	if err := json.Unmarshal(data, &sd); err != nil {
		return err
	}
	d.I18nCode = sd.I18nCode
	d.Value = sd.Value
	d.Unit = sd.Unit
	return nil
}
