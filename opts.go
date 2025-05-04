package redgiant

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type options struct {
	logger    zerolog.Logger
	localizer Localizer
}

type optFunc = func(*options)

func resolveOptions(optFuncs ...optFunc) *options {
	opts := &options{
		logger: log.Logger,
	}
	for _, fn := range optFuncs {
		fn(opts)
	}
	return opts
}

func WithLogger(l zerolog.Logger) optFunc {
	return func(opts *options) {
		opts.logger = l
	}
}

func WithLocalizer(l Localizer) optFunc {
	return func(opts *options) {
		opts.localizer = l
	}
}
