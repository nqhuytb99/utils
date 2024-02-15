package cache

import "sync"

type node struct {
	value      string
	prev, next *node
}

type lruQueue struct {
	head, tail *node
	size       int
	mu         *sync.Mutex
}

func (q *lruQueue) moveToFront(value string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	n := &node{value, nil, nil}

	if q.head == nil {
		q.head = n
		q.tail = n
	} else {
		n.next = q.head
		q.head.prev = n
		q.head = n
	}

	q.size++
}

func (q *lruQueue) removeFromTail() string {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.size == 0 {
		return ""
	}

	n := q.tail
	value := n.value

	if q.size == 1 {
		q.head = nil
		q.tail = nil
	} else {
		q.tail = n.prev
		q.tail.next = nil
		n.prev = nil
	}

	q.size--
	return value
}

// evictLRU evicts the least recently used item from the cache.
func (c *Cache[T]) evictLRU() {
	key := c.lruList.removeFromTail()

	c.locker.Lock()
	defer c.locker.Unlock()
	item, ok := c.items[keyFromString(key)]
	if !ok {
		return
	}

	c.size -= item.size
	delete(c.items, keyFromString(key))
}
