package worker

import (
	"context"
	"sync"
)

type WorkerPool[T any] struct {
	noWorkers   int
	wg          *sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	jobs        chan T
	processFunc func(T) error
	onFail      func(T, error)
}

func NewWorkerPool[T any](jobs chan T, processFunc func(T) error, onFail func(T, error), opts ...Option) *WorkerPool[T] {
	options := Options{
		noWorkers: DefaultNoWorkers,
	}

	for _, opt := range opts {
		opt.apply(&options)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool[T]{
		noWorkers:   options.noWorkers,
		jobs:        jobs,
		ctx:         ctx,
		cancel:      cancel,
		wg:          new(sync.WaitGroup),
		processFunc: processFunc,
		onFail:      onFail,
	}
}

func (p WorkerPool[T]) Run() {
	p.wg.Add(p.noWorkers)
	for i := 0; i < p.noWorkers; i++ {
		go p.do()
	}

	p.wg.Wait()
}

func (p WorkerPool[T]) Cancel() {
	p.cancel()
}

func (p WorkerPool[T]) do() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			return
		case j, ok := <-p.jobs:
			if !ok {
				return
			}

			err := p.processFunc(j)
			if err != nil {
				p.onFail(j, err)
			}
		}
	}
}
