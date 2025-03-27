package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pmeier/redgiant/internal/redgiant"
)

func apiRouteFuncs() []routeFunc {
	return []routeFunc{
		devices,
		summary,
	}
}

func devices(rg *redgiant.Redgiant) (string, string, echo.HandlerFunc) {
	return http.MethodGet, "/devices", func(c echo.Context) error {
		ds, err := rg.Devices()
		if err != nil {
			// TODO: error handling
			return err
		}
		return c.JSON(http.StatusOK, ds)
	}
}

func summary(rg *redgiant.Redgiant) (string, string, echo.HandlerFunc) {
	return http.MethodGet, "/summary", func(c echo.Context) error {
		type SummaryDevice struct {
			DeviceID int `query:"deviceID"`
		}
		var sd SummaryDevice
		err := c.Bind(&sd)
		if err != nil {
			// TODO: error handling
			return c.String(http.StatusBadRequest, "bad request")
		}

		s, err := rg.Summary(sd.DeviceID)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, s)
	}
}
