package server

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pmeier/redgiant"
	"github.com/rs/zerolog"
)

func redgiantWrapperRouteFunc[T any](path string, dataFunc func(*redgiant.Redgiant) (T, error)) routeFunc {
	return func(rg *redgiant.Redgiant, log zerolog.Logger) (string, string, echo.HandlerFunc) {
		return http.MethodGet, path, func(c echo.Context) error {
			d, err := dataFunc(rg)
			if err != nil {

				return err
			}
			return c.JSON(http.StatusOK, d)
		}
	}
}

func apiRouteFuncs() []routeFunc {
	return []routeFunc{
		redgiantWrapperRouteFunc("/about", (*redgiant.Redgiant).About),
		redgiantWrapperRouteFunc("/state", (*redgiant.Redgiant).State),
		redgiantWrapperRouteFunc("/devices", (*redgiant.Redgiant).Devices),
		liveData,
	}
}

func liveData(rg *redgiant.Redgiant, log zerolog.Logger) (string, string, echo.HandlerFunc) {
	return http.MethodGet, "/live-data/:deviceID", func(c echo.Context) error {
		type Params struct {
			DeviceID int    `param:"deviceID"`
			Services string `query:"services"`
			Language string `query:"lang"`
		}
		var p Params
		if err := c.Bind(&p); err != nil {
			return err
		}
		services := []string{}
		if p.Services != "" {
			services = strings.Split(p.Services, ",")
		}

		var d any
		var err error
		if p.Language == "" {
			d, err = rg.LiveData(p.DeviceID, services...)
		} else {
			d, err = rg.LocalizedLiveData(p.DeviceID, redgiant.GermanLanguage, services...)
		}
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, d)
	}
}
