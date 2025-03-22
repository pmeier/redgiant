package health

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

type HealthParams struct {
	RedgiantHost string
	RedgiantPort uint
}

func HealthRouteFunc() (string, string, echo.HandlerFunc) {
	return http.MethodGet, "/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}
}

func IsHealthy(host string, port uint) bool {
	r, err := http.Get((&url.URL{Scheme: "http",
		Host: fmt.Sprintf("%s:%d", host, port),
		Path: "/health"}).String())
	if err == nil && r.StatusCode == http.StatusOK {
		return true
	} else {
		return false
	}
}

func WaitForHealthy(host string, port uint, d time.Duration) error {
	timeout := time.After(d)
	for {
		select {
		case <-timeout:
			return errors.New("server failed to start")
		default:
			if IsHealthy(host, port) {
				return nil
			} else {
				<-time.After(time.Second)
			}
		}
	}
}

func Start(p HealthParams) {
	var c int
	if !IsHealthy(p.RedgiantHost, p.RedgiantPort) {
		c = 1
	}
	os.Exit(c)
}
