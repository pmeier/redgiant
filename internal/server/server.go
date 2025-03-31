package server

import (
	"embed"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pmeier/redgiant"
	"github.com/pmeier/redgiant/internal/health"
	"github.com/rs/zerolog"
)

//go:embed static/*
var staticFS embed.FS

//go:embed templates/*
var templatesFS embed.FS

type basicRouteFunc = func() (string, string, echo.HandlerFunc)
type routeFunc = func(*redgiant.Redgiant, zerolog.Logger) (string, string, echo.HandlerFunc)

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

type Server struct {
	*echo.Echo
	sp  ServerParams
	log zerolog.Logger
}

func newServer(sp ServerParams, rg *redgiant.Redgiant, logger zerolog.Logger) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.StaticFS("/static", echo.MustSubFS(staticFS, "static"))
	e.Renderer = newTemplate(templatesFS, "templates")

	routeFuncs := []routeFunc{
		wrapBasicRouteFunc(health.HealthRouteFunc),
		wrapBasicRouteFunc(redirectFuncFactory("/", "/ui/")),
	}
	routeFuncs = append(routeFuncs, withPrefix("/api", apiRouteFuncs()...)...)
	routeFuncs = append(routeFuncs, withPrefix("/ui", uiRouteFuncs()...)...)
	for _, routeFunc := range routeFuncs {
		method, path, handler := routeFunc(rg, logger)
		e.Add(method, path, handler)
	}

	return &Server{Echo: e, sp: sp}
}

func (s *Server) Start(timeout time.Duration) error {
	go func() {
		s.Echo.Start(fmt.Sprintf("%s:%d", s.sp.Host, s.sp.Port))
	}()

	if err := health.WaitForHealthy(s.sp.Host, s.sp.Port, timeout); err != nil {
		return err
	}

	s.log.Info().Str("host", s.sp.Host).Int("port", int(s.sp.Port)).Msg("server started")
	return nil
}

func redirectFuncFactory(src, dst string) basicRouteFunc {
	return func() (string, string, echo.HandlerFunc) {
		return http.MethodGet, src, func(c echo.Context) error {
			return c.Redirect(http.StatusMovedPermanently, dst)
		}
	}
}
