package cache

import "time"

// startGC starts a goroutine that periodically calls the collectGarbage method on the cache object.
func (c *Cache[T]) startGC(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			c.collectGarbage()
		}
	}()
}

// collectGarbage removes expired items from the cache.
func (c *Cache[T]) collectGarbage() {
	c.locker.Lock()
	defer c.locker.Unlock()

	now := time.Now()
	for key, entry := range c.items {
		if entry.exp.Before(now) {
			c.delete(key)
		}
	}
}
