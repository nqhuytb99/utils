package cache

type Options struct {
	capacity uint64
}

type CacheOption interface {
	apply(*Options)
}

type capacityOption uint64

func (o capacityOption) apply(opts *Options) {
	opts.capacity = uint64(o)
}

func WithCapacity(o uint64) CacheOption {
	return capacityOption(o)
}
