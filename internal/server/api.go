package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pmeier/redgiant"
	"github.com/rs/zerolog"
)

func apiRouteFuncs() []routeFunc {
	return []routeFunc{
		devices,
	}
}

func devices(rg *redgiant.Redgiant, log zerolog.Logger) (string, string, echo.HandlerFunc) {
	return http.MethodGet, "/devices", func(c echo.Context) error {
		ds, err := rg.Devices()
		if err != nil {
			// TODO: error handling
			return err
		}
		return c.JSON(http.StatusOK, ds)
	}
}
