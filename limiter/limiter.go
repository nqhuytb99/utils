// Package limiter provides a sliding window rate limiter that limits the number of requests allowed within a specified window of time.
package limiter

import (
	"sync"
	"time"
)

// Limiter represents a sliding window rate limiter.
type Limiter struct {
	windows   time.Duration
	threshold int
	// buffer is a circular buffer of the last threshold timestamps of requests.
	buffer *Ring[time.Time]
	mu     *sync.RWMutex
}

func NewLimiter(windows time.Duration, threshold int) *Limiter {
	return &Limiter{
		windows:   windows,
		threshold: threshold,
		buffer:    newRing[time.Time](threshold),
		mu:        new(sync.RWMutex),
	}
}

// AvailableAt calculates and returns the time at which the next request can be made.
func (l *Limiter) AvailableAt() time.Time {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.buffer.Value.Add(l.windows).Before(time.Now()) {
		return time.Now()
	}

	return l.buffer.Value.Add(l.windows)
}

// TillAvailable calculates and returns the duration until the next request is allowed.
func (l *Limiter) TillAvailable() time.Duration {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.buffer.next().Value.Add(l.windows).Before(time.Now()) {
		return 0
	}

	return time.Until(l.buffer.next().Value.Add(l.windows))
}

// Inc increments the counter and returns true if the request is allowed, or false if the request is denied.
func (l *Limiter) Inc() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.buffer.next().Value.Add(l.windows).After(time.Now()) {
		return false
	}

	now := time.Now()
	l.buffer = l.buffer.next()
	l.buffer.Value = now

	return true
}
