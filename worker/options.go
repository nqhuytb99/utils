package worker

import (
	"runtime"
)

type Option interface {
	apply(*Options)
}

type Options struct {
	noWorkers int
}

var DefaultNoWorkers = runtime.NumCPU()

type NoWorkerOption int

func (o NoWorkerOption) apply(options *Options) {
	options.noWorkers = int(o)
}

func WithNoWorkers(noWorkers int) Option {
	return NoWorkerOption(noWorkers)
}
