package queue

import (
	"sync"
	"time"
)

type Queue[T any] struct {
	data       []T
	locker     *sync.RWMutex
	pushSignal chan bool
	options    QueueOptions
	mem        int
}

func NewQueue[T any](options ...QueueOption) Queue[T] {
	opts := defaultOptions()
	for _, option := range options {
		option.apply(&opts)
	}

	q := Queue[T]{
		data:       []T{},
		options:    opts,
		pushSignal: make(chan bool),
		locker:     new(sync.RWMutex),
		mem:        0,
	}

	return q
}

func (q *Queue[T]) Enqueue(value T) {
	q.enqueue(value)

	if q.options.sizeLimit > 0 && len(q.data) >= q.options.sizeLimit {
		q.pushSignal <- true
	}
	if q.options.memoryLimit > 0 && q.mem >= q.options.memoryLimit {
		q.pushSignal <- true
	}
}

func (q *Queue[T]) enqueue(value T) {
	q.locker.Lock()
	defer q.locker.Unlock()

	q.data = append(q.data, value)
}

func (q *Queue[T]) EnqueueWithChannel(input chan T) {
	go func() {
		for value := range input {
			q.Enqueue(value)
		}
	}()
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
			case <-q.pushSignal:
			case <-time.After(q.options.tickerInterval):
			}

			values := q.dequeueAll()
			if len(values) == 0 {
				continue
			}

			out <- values
		}
	}()

	return out
}
