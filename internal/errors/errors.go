package errors

import (
	"net/http"

	"maps"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type HTTPDetail uint8

const (
	NoHTTPDetail HTTPDetail = iota
	MessageHTTPDetail
	ContextHTTPDetail
)

type Context map[string]any

type options struct {
	cause      error
	context    Context
	httpCode   int
	httpDetail HTTPDetail
}

type optFunc = func(*options)

func withCause(err error) optFunc {
	return func(o *options) {
		o.cause = err
	}
}

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

type RedgiantError struct {
	err        error
	cause      error
	context    map[string]any
	httpCode   int
	httpDetail HTTPDetail
}

func New(msg string, opts ...optFunc) *RedgiantError {
	o := options{
		httpCode:   http.StatusInternalServerError,
		httpDetail: MessageHTTPDetail,
	}
	for _, fn := range opts {
		fn(&o)
	}
	return &RedgiantError{
		err:        errors.New(msg),
		cause:      o.cause,
		context:    o.context,
		httpCode:   o.httpCode,
		httpDetail: o.httpDetail,
	}
}

func Wrap(err error, opts ...optFunc) error {
	if _, ok := err.(*RedgiantError); ok || err == nil {
		return nil
	}

	return New(err.Error(), append([]optFunc{withCause(err)}, opts...)...)
}

func (rge RedgiantError) Error() string {
	return rge.err.Error()
}

func (rge RedgiantError) MarshalZerologObject(e *zerolog.Event) {
	e.Str(zerolog.MessageFieldName, rge.Error())

	s := pkgerrors.MarshalStack(rge.err).([]map[string]string)
	// drop the frames that show the construction of the RedgiantError
	var o int
	if rge.cause == nil {
		o = 1
	} else {
		o = 2
	}
	e.Any(zerolog.ErrorStackFieldName, s[o:])

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
