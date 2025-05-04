package serve

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pmeier/redgiant"
)

func getRouteFunc[P any, O any](path string, bindFunc func(echo.Context) (P, error), outputFunc func(*redgiant.Redgiant, P) (O, error)) routeFunc {
	return func(s *Server) (string, string, echo.HandlerFunc) {
		return http.MethodGet, path, func(c echo.Context) error {
			p, err := bindFunc(c)
			if err != nil {
				return err
			}

			o, err := outputFunc(s.rg, p)
			if err != nil {
				return err
			}
			return c.JSON(http.StatusOK, o)
		}
	}
}

func noInputRouteFunc[T any](path string, noInputFunc func(*redgiant.Redgiant) (T, error)) routeFunc {
	type Params struct{}

	bindFunc := func(c echo.Context) (Params, error) {
		var p Params
		if err := c.Bind(&p); err != nil {
			return Params{}, err
		}
		return p, nil
	}

	outputFunc := func(rg *redgiant.Redgiant, p Params) (T, error) {
		return noInputFunc(rg)
	}

	return getRouteFunc(path, bindFunc, outputFunc)
}

func dataRouteFunc[T any](path string, dataFunc func(*redgiant.Redgiant, int, redgiant.Language, ...string) (T, error)) routeFunc {
	type Params struct {
		DeviceID int               `param:"deviceID"`
		Language redgiant.Language `query:"lang"`
		Services []string          `query:"service"`
	}

	bindFunc := func(c echo.Context) (Params, error) {
		var p Params
		if err := c.Bind(&p); err != nil {
			return Params{}, err
		}
		return p, nil
	}

	outputFunc := func(rg *redgiant.Redgiant, p Params) (T, error) {
		return dataFunc(rg, p.DeviceID, p.Language, p.Services...)
	}

	return getRouteFunc(path, bindFunc, outputFunc)

}

func apiRouteFuncs() []routeFunc {
	return []routeFunc{
		noInputRouteFunc("/about", (*redgiant.Redgiant).About),
		noInputRouteFunc("/state", (*redgiant.Redgiant).State),
		noInputRouteFunc("/devices", (*redgiant.Redgiant).Devices),
		dataRouteFunc("/data/:deviceID/real", (*redgiant.Redgiant).RealData),
		dataRouteFunc("/data/:deviceID/direct", (*redgiant.Redgiant).DirectData),
	}
}
