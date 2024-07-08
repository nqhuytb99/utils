package queue

import (
	"context"
	"runtime"
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

func NewQueue[T any](ctx context.Context, options ...QueueOption) *Queue[T] {
	opts := defaultOptions()
	for _, option := range options {
		option.apply(&opts)
	}

	q := &Queue[T]{
		ctx:        ctx,
		data:       []T{},
		options:    opts,
		pushSignal: make(chan bool, 1),
		locker:     new(sync.RWMutex),
		mem:        0,
		out:        make(chan []T),
	}

	runtime.SetFinalizer(q, func(q *Queue[T]) {
		q.Close()
	})

	go q.watchForSignal()
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
	q.locker.Lock()
	q.data = append(q.data, value)
	q.mem += size.Of(value)
	q.locker.Unlock()

	if q.options.sizeLimit > 0 && len(q.data) >= q.options.sizeLimit {
		q.pushSignal <- true
	}

	if q.options.memoryLimit > 0 && q.mem+size.Of(value) >= q.options.memoryLimit {
		q.pushSignal <- true
	}
}

func (q *Queue[T]) Close() {
	close(q.pushSignal)
	close(q.out)
}

func (q *Queue[T]) flush() {
	q.locker.Lock()
	defer q.locker.Unlock()

	if len(q.data) == 0 {
		return
	}

	q.out <- q.data

	q.data = nil
	q.mem = 0
}

func (q *Queue[T]) watchForSignal() {
	for {
		select {
		case <-q.ctx.Done():
			q.flush()
			close(q.out)
			close(q.pushSignal)
			return
		case <-q.pushSignal:
			q.flush()
		case <-time.After(q.options.tickerInterval):
			q.flush()
		}
	}
}

func (q *Queue[T]) Receive() chan []T {
	return q.out
}
