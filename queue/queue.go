package queue

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/DmitriyVTitov/size"
)

type Queue[T any] struct {
	ctx       context.Context
	data      []T
	locker    *sync.RWMutex
	options   QueueOptions
	mem       int
	out       chan []T
	lastFlush time.Time
}

func NewQueue[T any](ctx context.Context, options ...QueueOption) *Queue[T] {
	opts := defaultOptions()
	for _, option := range options {
		option.apply(&opts)
	}

	q := &Queue[T]{
		ctx:     ctx,
		data:    []T{},
		options: opts,
		locker:  new(sync.RWMutex),
		mem:     0,
		out:     make(chan []T),
	}

	runtime.SetFinalizer(q, func(q *Queue[T]) {
		q.Close()
	})

	go q.watchForFlush()
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
		q.flush()
	}

	if q.options.memoryLimit > 0 && q.mem+size.Of(value) >= q.options.memoryLimit {
		q.flush()
	}
}

func (q *Queue[T]) Close() {
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
	q.lastFlush = time.Now()
}

func (q *Queue[T]) watchForFlush() {
	for {
		if q.options.flushInterval > 0 && time.Since(q.lastFlush) > q.options.flushInterval {
		}
	}
}

func (q *Queue[T]) Receive() chan []T {
	return q.out
}
