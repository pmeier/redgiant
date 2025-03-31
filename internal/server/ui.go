package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pmeier/redgiant"
	"github.com/rs/zerolog"
)

func uiRouteFuncs() []routeFunc {
	return []routeFunc{
		indexView,
	}
}

type indexData struct {
	Summary redgiant.Summary
}

func indexView(rg *redgiant.Redgiant, log zerolog.Logger) (string, string, echo.HandlerFunc) {
	return http.MethodGet, "/", func(c echo.Context) error {
		// FIXME: don't harcode this
		s, err := rg.Summary(1)
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, "index.html", indexData{s})
	}
}
