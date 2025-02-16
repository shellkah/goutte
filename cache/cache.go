package goutte

import (
	"container/list"
	"sync"
	"time"
)

// Represents an item stored in the cache.
// The expiration field is set if a TTL is provided; a zero value indicates no expiration.
type entry[K comparable, V any] struct {
	key        K
	value      V
	expiration time.Time
}

// Thread-safe LRU cache.
type Cache[K comparable, V any] struct {
	capacity int                 // Maximum number of items in the cache.
	mu       sync.Mutex          // Mutex to guard concurrent access.
	ll       *list.List          // Doubly linked list to maintain usage order.
	cache    map[K]*list.Element // Map for fast lookups: key -> *list.Element.
}

// Creates a new LRU cache with a given capacity.
// K must be a comparable type (like string, int, etc.) and V can be any type.
func New[K comparable, V any](capacity int) *Cache[K, V] {
	if capacity <= 0 {
		panic("capacity must be greater than zero")
	}
	return &Cache[K, V]{
		capacity: capacity,
		ll:       list.New(),
		cache:    make(map[K]*list.Element),
	}
}

// Retrieves the value associated with the given key.
// If the entry has expired, it is removed and a not-found result is returned.
// Otherwise, the accessed item is moved to the front of the list (most recently used).
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.cache[key]; ok {
		ent := ele.Value.(*entry[K, V])
		if !ent.expiration.IsZero() && time.Now().After(ent.expiration) {
			c.ll.Remove(ele)
			delete(c.cache, key)
			var zero V
			return zero, false
		}
		c.ll.MoveToFront(ele)
		return ent.value, true
	}

	var zero V
	return zero, false
}

// Inserts or updates a key-value pair in the cache without a TTL.
func (c *Cache[K, V]) Put(key K, value V) {
	c.PutWithTTL(key, value, 0)
}

// Inserts or updates a key-value pair in the cache with an optional TTL.
// A positive ttl will cause the entry to expire after the given duration.
func (c *Cache[K, V]) PutWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	// Update existing key.
	if ele, ok := c.cache[key]; ok {
		ent := ele.Value.(*entry[K, V])
		ent.value = value
		ent.expiration = expiration
		c.ll.MoveToFront(ele)
		return
	}

	// Add new entry.
	ent := &entry[K, V]{key: key, value: value, expiration: expiration}
	ele := c.ll.PushFront(ent)
	c.cache[key] = ele

	// Evict the least recently used item if over capacity.
	if c.ll.Len() > c.capacity {
		c.removeOldest()
	}
}

// Evicts the least recently used item from the cache.
func (c *Cache[K, V]) removeOldest() {
	ele := c.ll.Back()
	if ele == nil {
		return
	}
	c.ll.Remove(ele)
	ent := ele.Value.(*entry[K, V])
	delete(c.cache, ent.key)
}

// Removes a key from the cache if it exists.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.cache[key]; ok {
		c.ll.Remove(ele)
		delete(c.cache, key)
	}
}

// Clears all entries from the cache.
func (c *Cache[K, V]) Dump() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ll.Init()
	c.cache = make(map[K]*list.Element)
}

// Dynamically adjusts the capacity of the cache.
// If the new capacity is smaller than the current number of items,
// it evicts the least recently used items until the cache size fits the new capacity.
func (c *Cache[K, V]) SetCapacity(newCapacity int) {
	if newCapacity <= 0 {
		panic("new capacity must be greater than zero")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.capacity = newCapacity
	// Evict least recently used items until the cache fits the new capacity.
	for c.ll.Len() > c.capacity {
		c.removeOldest()
	}
}
