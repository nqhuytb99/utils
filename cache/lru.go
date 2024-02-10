package cache

// evictLRevictLeastRecentlyUsedItems evicts the least recently used items from the cache.
func (c *Cache[T]) evictLRevictLeastRecentlyUsedItems() {
	c.locker.Lock()
	defer c.locker.Unlock()

	for c.size > c.capacity {
		var lruEntry *CacheEntry[T]
		var lruKey string
		for key, entry := range c.items {
			if lruEntry == nil || entry.lastAccess.Before(lruEntry.lastAccess) {
				lruEntry = entry
				lruKey = key
			}

		}

		if lruEntry != nil {
			c.size -= lruEntry.size

			delete(c.items, lruKey)
		}
	}
}
