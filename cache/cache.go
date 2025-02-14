package goutte

import (
	"container/list"
	"sync"
)

// Entry in the cache.
type entry[K comparable, V any] struct {
	key   K
	value V
}

// Generic, thread-safe LRU cache.
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

// Retrieve the value associated with the given key.
// It returns the value and a boolean indicating whether the key was found.
// The accessed item is moved to the front of the list (most recently used).
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry[K, V]).value, true
	}

	// Return the zero value of V if key is not found.
	var zero V
	return zero, false
}

// Inserts or updates a key-value pair in the cache.
// If the key already exists, its value is updated and the entry is moved to the front.
// If the cache exceeds its capacity, the least recently used item is evicted.
func (c *Cache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update existing key.
	if ele, ok := c.cache[key]; ok {
		ele.Value.(*entry[K, V]).value = value
		c.ll.MoveToFront(ele)
		return
	}

	// Add new entry.
	ele := c.ll.PushFront(&entry[K, V]{key, value})
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
