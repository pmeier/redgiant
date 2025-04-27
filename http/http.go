package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pmeier/redgiant"
)

type Client struct {
	c    *http.Client
	host string
}

func NewClient(host string, port uint) *Client {
	return &Client{c: &http.Client{}, host: fmt.Sprintf("%s:%d", host, port)}
}

func assertResponseSuccessful(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		return nil
	}

	var msg string
	if c, _ := io.ReadAll(r.Body); len(c) == 0 {
		msg = r.Status
	} else {
		msg = string(c)
	}
	return errors.New(msg)
}

func (hc *Client) Health() error {
	r, err := hc.c.Get((&url.URL{Scheme: "http", Host: hc.host, Path: "/health"}).String())
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return assertResponseSuccessful(r)
}

func (hc *Client) getAPI(endpoint string, query url.Values, v any) error {
	u := url.URL{Scheme: "http", Host: hc.host, Path: fmt.Sprintf("/api%s", endpoint)}
	u.RawQuery = query.Encode()

	r, err := hc.c.Get(u.String())
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if err := assertResponseSuccessful(r); err != nil {
		return err
	}

	return json.NewDecoder(r.Body).Decode(v)
}

func (hc *Client) About() (redgiant.About, error) {
	var a redgiant.About
	return a, hc.getAPI("/about", nil, &a)
}

func (hc *Client) State() (redgiant.State, error) {
	var s redgiant.State
	return s, hc.getAPI("/state", nil, &s)
}

func (hc *Client) Devices() ([]redgiant.Device, error) {
	var ds []redgiant.Device
	return ds, hc.getAPI("/devices", nil, &ds)
}

func dataEndpointQuery(dataType string, deviceID int, lang redgiant.Language, services []string) (string, url.Values) {
	e := fmt.Sprintf("/data/%d/%s", deviceID, dataType)
	q := url.Values{}
	q.Add("lang", lang.String())
	for _, s := range services {
		q.Add("service", s)
	}
	return e, q
}

func (hc *Client) RealData(deviceID int, lang redgiant.Language, services ...string) ([]redgiant.RealMeasurement, error) {
	endpoint, q := dataEndpointQuery("real", deviceID, lang, services)
	var rms []redgiant.RealMeasurement
	return rms, hc.getAPI(endpoint, q, &rms)
}

func (hc *Client) DirectData(deviceID int, lang redgiant.Language, services ...string) ([]redgiant.DirectMeasurement, error) {
	endpoint, q := dataEndpointQuery("direct", deviceID, lang, services)
	var dms []redgiant.DirectMeasurement
	return dms, hc.getAPI(endpoint, q, &dms)
}
