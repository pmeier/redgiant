package redgiant

import "fmt"

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
	TotalFaults         int     `json:"total_fault,string"`
	TotalAlarms         int     `json:"total_alarm,string"`
	WirelessConnection  IntBool `json:"wireless_conn_sts,string"`
	WifiConnection      IntBool `json:"wifi_conn_sts,string"`
	Ethernet1Connection IntBool `json:"eth_conn_sts,string"`
	Ethernet2Connection IntBool `json:"eth2_conn_sts,string"`
	CloudConnection     IntBool `json:"cloud_conn_sts,string"`
}

// FIXME: add available services to the device
type Device struct {
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

type Datapoint struct {
	I18nCode string `json:"data_name"`
	Name     string
	Value    string `json:"data_value"`
	Unit     string `json:"data_unit"`
}
