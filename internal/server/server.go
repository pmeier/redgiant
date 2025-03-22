package server

import (
	"embed"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pmeier/redgiant/internal/health"
	"github.com/pmeier/redgiant/internal/redgiant"
	log "github.com/sirupsen/logrus"
)

//go:embed static/*
var staticFS embed.FS

//go:embed templates/*
var templatesFS embed.FS

type routeFunc = func(*redgiant.Redgiant, redgiant.Device) (string, string, echo.HandlerFunc)

type Server struct {
	*echo.Echo
	sp ServerParams
}

func newServer(sp ServerParams, rg *redgiant.Redgiant, device redgiant.Device) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.StaticFS("/static", echo.MustSubFS(staticFS, "static"))
	e.Renderer = newTemplate(templatesFS, "templates")

	routeFuncs := []routeFunc{
		func(*redgiant.Redgiant, redgiant.Device) (string, string, echo.HandlerFunc) {
			return health.HealthRouteFunc()
		},
		indexView,
	}
	for _, routeFunc := range routeFuncs {
		method, path, handler := routeFunc(rg, device)
		e.Add(method, path, handler)
	}

	return &Server{Echo: e, sp: sp}
}

func (s *Server) Start(timeout time.Duration) error {
	go func() {
		s.Echo.Start(fmt.Sprintf("%s:%d", s.sp.Host, s.sp.Port))
	}()

	return health.WaitForHealthy(s.sp.Host, s.sp.Port, timeout)
}

type indexData struct {
	Summary redgiant.Summary
}

func indexView(rg *redgiant.Redgiant, device redgiant.Device) (string, string, echo.HandlerFunc) {
	return http.MethodGet, "/", func(c echo.Context) error {
		s, err := rg.Summary(device)
		if err != nil {
			log.WithError(err)
			code := http.StatusInternalServerError
			c.String(code, http.StatusText(code))
		}
		return c.Render(http.StatusOK, "index.html", indexData{s})
	}
}
