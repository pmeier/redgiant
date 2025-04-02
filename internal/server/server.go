package server

import (
	"embed"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pmeier/redgiant"
	"github.com/pmeier/redgiant/internal/health"
	"github.com/rs/zerolog"
)

//go:embed static/*
var staticFS embed.FS

type routeFunc = func(*redgiant.Redgiant, zerolog.Logger) (string, string, echo.HandlerFunc)

type Server struct {
	*echo.Echo
	sp  ServerParams
	log zerolog.Logger
}

func newServer(sp ServerParams, rg *redgiant.Redgiant, logger zerolog.Logger) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Debug = true

	routeFuncs := []routeFunc{
		wrapBasicRouteFunc(health.HealthRouteFunc),
	}
	routeFuncs = append(routeFuncs, withPrefix("/api", apiRouteFuncs()...)...)
	for _, routeFunc := range routeFuncs {
		method, path, handler := routeFunc(rg, logger)
		e.Add(method, path, handler)
	}

	e.StaticFS("/", echo.MustSubFS(staticFS, "static"))

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogRemoteIP: true,
		LogURI:      true,
		LogStatus:   true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info().
				Str("origin", v.RemoteIP).
				Str("path", v.URI).
				Int("status_code", v.Status).
				Msg("request")

			return nil
		},
	}))

	return &Server{Echo: e, sp: sp, log: logger}
}

func (s *Server) Start(timeout time.Duration) error {
	log := s.log.With().Str("host", s.sp.Host).Int("port", int(s.sp.Port)).Logger()
	log.Info().Msg("starting")

	go func() {
		s.Echo.Start(fmt.Sprintf("%s:%d", s.sp.Host, s.sp.Port))
	}()

	if err := health.WaitForHealthy(s.sp.Host, s.sp.Port, timeout); err != nil {
		return err
	}

	log.Info().Msg("started")
	return nil
}

func wrapBasicRouteFunc(basicRouteFunc func() (string, string, echo.HandlerFunc)) routeFunc {
	return func(*redgiant.Redgiant, zerolog.Logger) (string, string, echo.HandlerFunc) {
		return basicRouteFunc()
	}
}

func withPrefix(prefix string, rfs ...routeFunc) []routeFunc {
	prfs := make([]routeFunc, 0, len(rfs))
	for _, rf := range rfs {
		prfs = append(prfs, func(rg *redgiant.Redgiant, log zerolog.Logger) (string, string, echo.HandlerFunc) {
			method, route, handlerFunc := rf(rg, log)
			return method, prefix + route, handlerFunc
		})
	}
	return prfs
}
