package http

import (
	"redgiant"
	"url"
)

type HTTPClient struct {
	c *http.Client
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{}
}

func (hc *HTTPClient) Health() error {
	return nil
}


func (hc *HTTPClient) getAPI(endpoint string, query url.Values, v any) error {
	u := url.URL{Scheme: "http", Host: "FIXME", Path: fmt.Sprintf("/api%s", ednpoint)}
	u.RawQuery = query.Encode()

	r, err := hc.c.Get(u)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if !(r.StatusCode >= 200 && r.StatusCode < 300) {
		if c, _ := io.ReadAll(); len(c) == 0 {
			msg = r.Status
		} else {
			msg = string(c)
		}
		return errors.New(msg)
	}

	return json.NewDecoder(r.Body).Decode(v)
}

func (hc *HTTPClient) About() (redgiant.About, error) {
	var a redgiant.About
	return a, hc.getAPI("/about", nil, &a)
}

func (hc *HTTPClient) State() (redgiant.State, error) {
	var s redgiant.State
	return s, hc.getAPI("/state", nil, &s)
}

func (hc *HTTPClient) Devices() ([]redgiant.Device, error) {
	var ds []redgiant.Device
	return ds, hc.getAPI("/devices", nil, &ds)
}

func dataEndpointQuery(dataType string, deviceID int, lang Language, services ...string) (string, url.Values) {
	e := fmt.Sprintf("/data/%d/%s", deviceID, dataType)
	q := url.Values{}
	q.Add("lang", lang.String())
	for _, s := range services {
		q.Add("service", s)
	}
	return e, q
}

func (hc *HTTPClient) RealData(deviceID int, lang Language, services ...string) ([]redgiant.RealMeasurement, error) {
	endpoint, q := dataEndpointQuery("real", deviceID, lang, services)
	var rms []redgiant.RealMeasurement
	return rms, hc.getAPI(, q, &rms)
}

func (hc *HTTPClient) DirectData(deviceID int, lang Language, services ...string) ([]redgiant.DirectMeasurement, error) {
	endpoint, q := dataEndpointQuery("direct", deviceID, lang, services)
	var dms []redgiant.DirectMeasurement
	return dms, hc.getAPI(, q, &dms)
}
