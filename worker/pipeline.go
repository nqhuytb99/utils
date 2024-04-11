package worker

import (
	"context"
)

type Pipeline[T any, Y any] struct {
	noWorkers   int
	jobs        chan T
	ctx         context.Context
	cancel      context.CancelFunc
	processFunc func(T) (Y, error)
	onFail      func(T, error)
}

func NewPipeline[T any, Y any](jobs chan T, processFunc func(T) (Y, error), onFail func(T, error), opts ...Option) *Pipeline[T, Y] {
	options := Options{
		noWorkers: DefaultNoWorkers,
	}

	for _, opt := range opts {
		opt.apply(&options)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Pipeline[T, Y]{
		noWorkers:   options.noWorkers,
		jobs:        jobs,
		ctx:         ctx,
		cancel:      cancel,
		processFunc: processFunc,
		onFail:      onFail,
	}
}

func (p Pipeline[T, Y]) Run() chan Y {
	results := make([]chan Y, p.noWorkers)
	for i := 0; i < p.noWorkers; i++ {
		results[i] = p.do()
	}

	return Merge(results...)
}

func (p Pipeline[T, Y]) Cancel() {
	p.cancel()
}

func (p Pipeline[T, Y]) do() chan Y {
	out := make(chan Y)
	go func() {
		defer close(out)
		for {
			select {
			case <-p.ctx.Done():
				return
			case j, ok := <-p.jobs:
				if !ok {
					return
				}

				result, err := p.processFunc(j)
				if err != nil {
					p.onFail(j, err)
					continue
				}

				out <- result
			}
		}
	}()

	return out
}
