package redgiant

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type commonOpts struct {
	logger zerolog.Logger
}

type commonOptFunc = func(*commonOpts)

func defaultCommonOpts() *commonOpts {
	return &commonOpts{logger: log.Logger}
}

func resolveCommonOpts(optFuncs []commonOptFunc) *commonOpts {
	o := defaultCommonOpts()
	for _, fn := range optFuncs {
		fn(o)
	}
	return o
}

func WithLogger(l zerolog.Logger) commonOptFunc {
	return func(opts *commonOpts) {
		opts.logger = l
	}
}
