package queue

import (
	"time"
)

type QueueOption interface {
	apply(*QueueOptions)
}

type QueueOptions struct {
	tickerInterval time.Duration
	sizeLimit      int
	memoryLimit    int
}

type tickerIntervalOption struct {
	value time.Duration
}

func (o *tickerIntervalOption) apply(options *QueueOptions) {
	options.tickerInterval = o.value
}

func WithTickerInterval(value time.Duration) QueueOption {
	return &tickerIntervalOption{value: value}
}

type sizeLimitOption struct {
	value int
}

func (o *sizeLimitOption) apply(options *QueueOptions) {
	options.sizeLimit = o.value
}

func WithSizeLimit(value int) QueueOption {
	return &sizeLimitOption{value: value}
}

type memoryLimitOption struct {
	value int
}

func (o *memoryLimitOption) apply(options *QueueOptions) {
	options.memoryLimit = o.value
}

func WithMemoryLimit(value int) QueueOption {
	return &memoryLimitOption{value: value}
}

func defaultOptions() QueueOptions {
	return QueueOptions{
		tickerInterval: 10 * time.Second,
		sizeLimit:      0,
		memoryLimit:    0,
	}
}
