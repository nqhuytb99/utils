// Package cache provides a type-safe in-memory cache implementation.
package cache

import (
	"hash/fnv"
	"maps"
	"runtime"
	"sync"
	"time"

	"github.com/DmitriyVTitov/size"
)

const (
	defaultGCInverval = 2 * time.Minute
	timeSize          = 24
	mapReferenceSize  = 24
	uint64Size        = 8
	rwMutexSize       = 24
	cacheMapSize      = mapReferenceSize + 2*uint64Size + rwMutexSize
)

var defaultCapacity uint64

func init() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// Calculate the maximum capacity based on 12.5% of total memory, capped at 200 MB
	defaultCapacity = min(m.Sys<<10>>3, 200*1000*1000)
}

// CacheEntry holds the cached data and its expiration time
type CacheEntry[T any] struct {
	data T
	exp  time.Time
	size uint64
}

// Cache implements a type-safe in-memory cache
type Cache[T any] struct {
	items  map[uint64]*CacheEntry[T]
	locker *sync.RWMutex

	lruList *lruQueue
	size    uint64
	Options
}

// entrySize calculates the size of an entry based on the key and data
func entrySize[T any](key string, data T) uint64 {
	keySize := uint64(len(key))
	dataSize := uint64(size.Of(data))
	totalSize := keySize + dataSize + 2*timeSize + uint64Size
	return totalSize
}

// New creates a new cache instance with the provided options.
func New[T any](options ...CacheOption) *Cache[T] {
	o := Options{
		capacity: defaultCapacity,
	}
	for _, option := range options {
		option.apply(&o)
	}
	c := &Cache[T]{
		items:  make(map[uint64]*CacheEntry[T]),
		locker: new(sync.RWMutex),
		size:   cacheMapSize,
		lruList: &lruQueue{
			mu: new(sync.Mutex),
		},
		Options: o,
	}
	c.startGC(defaultGCInverval)
	return c
}

// Get retrieves the value associated with the given key from the cache.
// If the key is not found or if the entry has expired, it returns the zero value of type T and false.
// Otherwise, it returns the value and true.
func (c *Cache[T]) Get(key string) (T, bool) {
	c.locker.Lock()
	defer c.locker.Unlock()
	hashKey := keyFromString(key)
	entry, ok := c.items[hashKey]
	if !ok || entry.exp.Before(time.Now()) {
		delete(c.items, hashKey)
		return zero[T](), false
	}

	c.lruList.moveToFront(key)
	return entry.data, true
}

// Set adds or updates a cache entry with the given key, data, and expiration duration.
func (c *Cache[T]) Set(key string, data T, exp time.Duration) {
	newSize := entrySize[T](key, data)

	c.incSize(newSize)
	for c.size > c.capacity {
		c.evictLRU()
	}

	c.set(keyFromString(key), &CacheEntry[T]{
		data: data,
		exp:  time.Now().Add(exp),
		size: newSize,
	})
}

// Filter applies a filter function to the cache entries and returns a slice of filtered values.
func (c *Cache[T]) Filter(fn func(T) bool) []T {
	snapshot := c.snapshot()
	var result []T
	for _, entry := range snapshot {
		if fn(entry.data) {
			result = append(result, entry.data)
		}
	}
	return result
}

// DeleteMatchingEntries deletes cache entries that match the given filter function.
func (c *Cache[T]) DeleteMatchingEntries(fn func(T) bool) {
	c.locker.Lock()
	defer c.locker.Unlock()

	for key, entry := range c.items {
		if fn(entry.data) {
			delete(c.items, key)
		}
	}
}

// snapshot creates a snapshot of the cache items.
func (c *Cache[T]) snapshot() map[uint64]*CacheEntry[T] {
	c.locker.Lock()
	defer c.locker.Unlock()

	return maps.Clone(c.items)
}

func (c *Cache[T]) set(key uint64, entry *CacheEntry[T]) {
	c.locker.Lock()
	defer c.locker.Unlock()

	c.items[key] = entry
}

// incSize increments the size of the cache
func (c *Cache[T]) incSize(size uint64) {
	c.locker.Lock()
	defer c.locker.Unlock()

	c.size += size
}

// Delete removes the data associated with a key
func (c *Cache[T]) Delete(key string) {
	c.locker.Lock()
	defer c.locker.Unlock()

	delete(c.items, keyFromString(key))
}

func (c *Cache[T]) delete(hashKey uint64) {
	c.locker.Lock()
	defer c.locker.Unlock()

	delete(c.items, hashKey)
}

// Prune removes all cache's items
func (c *Cache[T]) Prune(key string) {
	c.locker.Lock()
	defer c.locker.Unlock()

	c.items = make(map[uint64]*CacheEntry[T])
	c.size = cacheMapSize
}

// zero[T] returns the zero value of type T
func zero[T any]() T {
	var z T
	return z
}

func (c *Cache[T]) Size() uint64 {
	return c.size
}

func keyFromString(key string) (hashKey uint64) {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}
