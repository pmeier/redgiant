package errors

import (
	"bytes"
	"net/http"

	"github.com/rs/zerolog"
)

type Error struct {
	StatusCode int
	Redacted   bool
	e          *zerolog.Event
	msg        string
	b          *bytes.Buffer
	text       string
}

func New(msg string) *Error {
	var b bytes.Buffer
	l := zerolog.New(&b)
	return &Error{StatusCode: http.StatusInternalServerError, Redacted: true, e: l.Log(), msg: msg, b: &b}
}

func (e *Error) Error() string {
	if e.text == "" {
		e.e.Msg(e.msg)
		e.text = e.b.String()
	}

	return e.text
}

func (err *Error) HTTPStatusCode(statusCode int) *Error {
	err.StatusCode = statusCode
	return err
}

func (err *Error) HTTPRedacted(redacted bool) *Error {
	err.Redacted = redacted
	return err
}
