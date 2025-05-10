package serve

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pmeier/redgiant"
	"github.com/pmeier/redgiant/internal/errors"
	"github.com/pmeier/redgiant/internal/health"
	"github.com/rs/zerolog"
)

type Server struct {
	*echo.Echo
	rg  *redgiant.Redgiant
	log zerolog.Logger
}

type routeFunc = func(*Server) (string, string, echo.HandlerFunc)

//go:embed static/*
var staticFS embed.FS

func newServer(rg *redgiant.Redgiant, logger zerolog.Logger) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	s := &Server{Echo: e, rg: rg, log: logger}

	routeFuncs := []routeFunc{
		wrapBasicRouteFunc(health.HealthRouteFunc),
	}
	routeFuncs = append(routeFuncs, withPrefix("/api", apiRouteFuncs()...)...)
	for _, routeFunc := range routeFuncs {
		method, path, handler := routeFunc(s)
		e.Add(method, path, handler)
	}

	e.StaticFS("/", echo.MustSubFS(staticFS, "static"))

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper: func(c echo.Context) bool {
			return c.Request().URL.Path == "/health" && logger.GetLevel() > zerolog.DebugLevel
		},
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

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		var code int
		var logData map[string]any
		var httpData map[string]any
		if rgerr, ok := err.(*errors.Error); ok {
			code = rgerr.StatusCode
			if err := json.Unmarshal([]byte(err.Error()), &logData); err != nil {
				panic(err.Error())
			}
			if !rgerr.Redacted {
				httpData = logData
			}
		} else {
			code = http.StatusInternalServerError
		}

		if httpData == nil {
			httpData = map[string]any{zerolog.MessageFieldName: http.StatusText(code)}
		}

		e := logger.Error()
		for k, v := range logData {
			e.Any(k, v)
		}
		e.Send()

		c.JSON(code, httpData)
	}

	return s
}

func (s *Server) Start(host string, port uint, timeout time.Duration) error {
	log := s.log.With().Str("host", host).Int("port", int(port)).Logger()
	log.Info().Msg("starting")

	go func() {
		s.Echo.Start(fmt.Sprintf("%s:%d", host, port))
	}()

	if err := health.WaitForHealthy(host, port, timeout); err != nil {
		return err
	}

	log.Info().Msg("started")
	return nil
}

func wrapBasicRouteFunc(basicRouteFunc func() (string, string, echo.HandlerFunc)) routeFunc {
	return func(*Server) (string, string, echo.HandlerFunc) {
		return basicRouteFunc()
	}
}

func withPrefix(prefix string, rfs ...routeFunc) []routeFunc {
	prfs := make([]routeFunc, 0, len(rfs))
	for _, rf := range rfs {
		prfs = append(prfs, func(s *Server) (string, string, echo.HandlerFunc) {
			method, route, handlerFunc := rf(s)
			return method, prefix + route, handlerFunc
		})
	}
	return prfs
}
