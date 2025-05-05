package redgiant

import (
	"net/http"

	"github.com/rs/zerolog"
)

type Options struct {
	Logger     zerolog.Logger
	Localizer  Localizer
	HTTPClient *http.Client
}

type OptFunc = func(*Options)

func ResolveOptions(optFuncs ...OptFunc) *Options {
	opts := &Options{}
	for _, fn := range optFuncs {
		fn(opts)
	}
	return opts
}

func WithLogger(l zerolog.Logger) OptFunc {
	return func(opts *Options) {
		opts.Logger = l
	}
}

func WithLocalizer(l Localizer) OptFunc {
	return func(opts *Options) {
		opts.Localizer = l
	}
}

func WithHTTPClient(c *http.Client) OptFunc {
	return func(opts *Options) {
		opts.HTTPClient = c
	}
}
