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
	out        chan []T
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
		out:        make(chan []T),
	}

	q.watchForSignal()
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

func (q *Queue[T]) Close() {
	close(q.pushSignal)
}

func (q *Queue[T]) dequeueAll() {
	q.locker.Lock()
	defer q.locker.Unlock()

	if len(q.data) == 0 {
		return
	}

	q.out <- q.data

	q.data = []T{}
	q.mem = 0
}

func (q *Queue[T]) watchForSignal() {
	go func() {
		for {
			select {
			case <-q.ctx.Done():
				q.dequeueAll()
				close(q.out)
				return
			case <-q.pushSignal:
				q.dequeueAll()
			case <-time.After(q.options.tickerInterval):
				q.dequeueAll()
			}
		}
	}()
}

func (q *Queue[T]) Receive() chan []T {
	return q.out
}
