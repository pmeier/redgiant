package errors

import (
	"net/http"

	"maps"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type RedgiantErrorer interface {
	error
	zerolog.LogObjectMarshaler
	SendAsResponse(c echo.Context)
}

type HTTPDetail uint8

const (
	NoHTTPDetail HTTPDetail = iota
	MessageHTTPDetail
	ContextHTTPDetail
)

type Context map[string]any

type options struct {
	context      Context
	httpCode     int
	httpDetail   HTTPDetail
	hiddenFrames uint
}

type optFunc = func(*options)

func WithContext(ctx Context) optFunc {
	return func(o *options) {
		o.context = ctx
	}
}

func WithHTTPCode(code int) optFunc {
	return func(o *options) {
		o.httpCode = code
	}
}

func WithHTTPDetail(detail HTTPDetail) optFunc {
	return func(o *options) {
		o.httpDetail = detail
	}
}

func WithHiddenFrames(n uint) optFunc {
	return func(o *options) {
		o.hiddenFrames = n
	}
}

type RedgiantError struct {
	err          error
	context      map[string]any
	httpCode     int
	httpDetail   HTTPDetail
	hiddenFrames uint
}

func New(msg string, opts ...optFunc) *RedgiantError {
	o := options{
		httpCode:     http.StatusInternalServerError,
		httpDetail:   MessageHTTPDetail,
		hiddenFrames: 1,
	}
	for _, fn := range opts {
		fn(&o)
	}
	return &RedgiantError{
		err:          errors.New(msg),
		context:      o.context,
		httpCode:     o.httpCode,
		httpDetail:   o.httpDetail,
		hiddenFrames: o.hiddenFrames,
	}
}

func Wrap(err error, opts ...optFunc) error {
	if _, ok := err.(*RedgiantError); ok || err == nil {
		return nil
	}

	return New(err.Error(), append([]optFunc{WithHiddenFrames(2)}, opts...)...)
}

func (rge RedgiantError) Error() string {
	return rge.err.Error()
}

func (rge RedgiantError) MarshalZerologObject(e *zerolog.Event) {
	e.Str(zerolog.MessageFieldName, rge.Error())

	s := pkgerrors.MarshalStack(rge.err).([]map[string]string)
	// drop the frames that show the construction of the RedgiantError
	e.Any(zerolog.ErrorStackFieldName, s[rge.hiddenFrames:])

	for k, v := range rge.context {
		e.Any(k, v)
	}
}

func (rge RedgiantError) SendAsResponse(c echo.Context) {
	var m string
	if rge.httpDetail == NoHTTPDetail {
		m = http.StatusText(rge.httpCode)
	} else {
		m = rge.Error()
	}
	e := map[string]any{zerolog.MessageFieldName: m}
	maps.Copy(e, rge.context)
	i := map[string]any{zerolog.ErrorFieldName: e}
	c.JSON(rge.httpCode, &i)
}
