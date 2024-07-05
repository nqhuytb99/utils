package queue

import (
	"context"
	"sync"
	"time"

	"github.com/DmitriyVTitov/size"
)

type Queue[T any] struct {
	ctx        context.Context
	data       []T
	locker     *sync.RWMutex
	pushSignal chan bool
	options    QueueOptions
	mem        int
}

func NewQueue[T any](ctx context.Context, options ...QueueOption) Queue[T] {
	opts := defaultOptions()
	for _, option := range options {
		option.apply(&opts)
	}

	q := Queue[T]{
		ctx:        ctx,
		data:       []T{},
		options:    opts,
		pushSignal: make(chan bool),
		locker:     new(sync.RWMutex),
		mem:        0,
	}

	go func() {
		ticker := time.NewTicker(q.options.tickerInterval)
		defer ticker.Stop()

		for range ticker.C {
			q.pushSignal <- true
		}
	}()

	return q
}

func (q *Queue[T]) Enqueue(value T) error {
	select {
	case <-q.ctx.Done():
		return q.ctx.Err()
	default:
		q.enqueue(value)
	}
	return nil
}

func (q *Queue[T]) enqueue(value T) {
	if q.options.sizeLimit > 0 && len(q.data) > q.options.sizeLimit {
		q.pushSignal <- true
	}

	if q.options.memoryLimit > 0 && q.mem+size.Of(value) > q.options.memoryLimit {
		q.pushSignal <- true
	}

	q.locker.Lock()
	defer q.locker.Unlock()
	q.data = append(q.data, value)
	q.mem += size.Of(value)
}

func (q *Queue[T]) EnqueueWithChannel(input chan T) {
	go func() {
		for value := range input {
			q.Enqueue(value)
		}

		close(q.pushSignal)
	}()
}

func (q *Queue[T]) Close() {
	close(q.pushSignal)
}

func (q *Queue[T]) dequeueAll() []T {
	q.locker.Lock()
	defer q.locker.Unlock()

	values := q.data

	q.data = []T{}
	q.mem = 0

	return values
}

func (q *Queue[T]) Receive() chan []T {
	out := make(chan []T)

	go func() {
		defer close(out)
		for {
			select {
			case <-q.ctx.Done():
				values := q.dequeueAll()
				if len(values) == 0 {
					continue
				}

				out <- values
				return
			case <-q.pushSignal:
				values := q.dequeueAll()
				if len(values) == 0 {
					continue
				}

				out <- values
			}
		}
	}()

	return out
}
