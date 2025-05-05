package http

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pmeier/redgiant"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Redgiant struct {
	host string
	c    *http.Client
	log  zerolog.Logger
}

func NewRedgiant(host string, port uint, opts ...redgiant.OptFunc) *Redgiant {
	o := redgiant.ResolveOptions(append([]redgiant.OptFunc{
		redgiant.WithLogger(log.Logger),
		redgiant.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: time.Second * 60,
		}),
	}, opts...)...)
	return &Redgiant{host: fmt.Sprintf("%s:%d", host, port), c: o.HTTPClient, log: o.Logger}
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

func (rg *Redgiant) Health() error {
	rg.log.Trace().Msg("Redgiant.Health()")

	r, err := rg.c.Get((&url.URL{Scheme: "http", Host: rg.host, Path: "/health"}).String())
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return assertResponseSuccessful(r)
}

func (rg *Redgiant) getAPI(endpoint string, query url.Values, v any) error {
	rg.log.Trace().Str("endpoint", endpoint).Func(func(e *zerolog.Event) { e.Str("query", query.Encode()) }).Msg("Redgiant.getAPI()")

	u := url.URL{Scheme: "http", Host: rg.host, Path: fmt.Sprintf("/api%s", endpoint)}
	u.RawQuery = query.Encode()

	rg.log.Debug().Func(func(e *zerolog.Event) { e.Str("url", u.String()) }).Msg("GET")
	r, err := rg.c.Get(u.String())
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if err := assertResponseSuccessful(r); err != nil {
		return err
	}

	return json.NewDecoder(r.Body).Decode(v)
}

func (rg *Redgiant) About() (redgiant.About, error) {
	rg.log.Trace().Msg("Redgiant.About()")

	var a redgiant.About
	return a, rg.getAPI("/about", nil, &a)
}

func (rg *Redgiant) State() (redgiant.State, error) {
	rg.log.Trace().Msg("Redgiant.State()")

	var s redgiant.State
	return s, rg.getAPI("/state", nil, &s)
}

func (rg *Redgiant) Devices() ([]redgiant.Device, error) {
	rg.log.Trace().Msg("Redgiant.Devices()")

	var ds []redgiant.Device
	return ds, rg.getAPI("/devices", nil, &ds)
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

func (rg *Redgiant) RealData(deviceID int, lang redgiant.Language, services ...string) ([]redgiant.RealMeasurement, error) {
	rg.log.Trace().Int("deviceID", deviceID).Stringer("lang", lang).Strs("services", services).Msg("Redgiant.RealData()")

	endpoint, q := dataEndpointQuery("real", deviceID, lang, services)
	var rms []redgiant.RealMeasurement
	return rms, rg.getAPI(endpoint, q, &rms)
}

func (rg *Redgiant) DirectData(deviceID int, lang redgiant.Language, services ...string) ([]redgiant.DirectMeasurement, error) {
	rg.log.Trace().Int("deviceID", deviceID).Stringer("lang", lang).Strs("services", services).Msg("Redgiant.DirectData()")

	endpoint, q := dataEndpointQuery("direct", deviceID, lang, services)
	var dms []redgiant.DirectMeasurement
	return dms, rg.getAPI(endpoint, q, &dms)
}
