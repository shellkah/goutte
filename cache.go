package goutte

import (
	"container/heap"
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
	exp        *expEntry[K]
}

// Thread-safe & type-safe LRU cache.
type Cache[K comparable, V any] struct {
	capacity int                 // maximum number of items in the cache
	mu       sync.Mutex          // guards cache and ll below
	ll       *list.List          // doubly-linked list for LRU ordering
	cache    map[K]*list.Element // map from key to list element

	// Fields for TTL expiration management:
	expHeap  expHeap[K]    // min-heap of expiration entries
	updateCh chan struct{} // signals that a new expiration might be sooner
	done     chan struct{} // closed when the cache is shutting down
}

// Creates a new LRU cache with a given capacity.
// K must be a comparable type (like string, int, etc.) and V can be any type.
func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	if capacity <= 0 {
		panic("capacity must be greater than zero")
	}
	c := &Cache[K, V]{
		capacity: capacity,
		ll:       list.New(),
		cache:    make(map[K]*list.Element),
		updateCh: make(chan struct{}, 1),
		done:     make(chan struct{}),
	}
	heap.Init(&c.expHeap)
	go c.expirationProcessor()
	return c
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
func (c *Cache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, 0)
}

// Inserts or updates a key-value pair in the cache with an optional TTL.
// A positive ttl will cause the entry to expire after the given duration.
func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Update existing key.
	if ele, ok := c.cache[key]; ok {
		ent := ele.Value.(*entry[K, V])
		ent.value = value
		ent.expiration = expiration
		c.ll.MoveToFront(ele)

		if ttl > 0 {
			if ent.exp != nil {
				// Update existing expiration entry.
				ent.exp.expiration = expiration
				heap.Fix(&c.expHeap, ent.exp.index)
			} else {
				// Create a new expiration entry and attach it.
				expE := &expEntry[K]{key: key, expiration: expiration}
				ent.exp = expE
				heap.Push(&c.expHeap, expE)
			}
			c.signalExpirationUpdate()
		} else {
			// TTL is 0: cancel any existing expiration.
			if ent.exp != nil {
				ent.exp.canceled = true
				ent.exp = nil
			}
		}
		return
	}

	// Add new entry.
	ent := &entry[K, V]{key: key, value: value, expiration: expiration}
	ele := c.ll.PushFront(ent)
	c.cache[key] = ele

	// If the item has a TTL, attach an expiration entry.
	if ttl > 0 {
		expE := &expEntry[K]{key: key, expiration: expiration}
		ent.exp = expE
		heap.Push(&c.expHeap, expE)
		c.signalExpirationUpdate()
	}

	// Evict the least recently used item if over capacity.
	if c.ll.Len() > c.capacity {
		c.removeOldestLocked()
	}
}

func (c *Cache[K, V]) signalExpirationUpdate() {
	select {
	case c.updateCh <- struct{}{}:
	default:
		// already a signal in the channel; no need to block
	}
}

func (c *Cache[K, V]) removeOldestLocked() {
	ele := c.ll.Back()
	if ele == nil {
		return
	}
	ent := ele.Value.(*entry[K, V])
	if ent.exp != nil {
		ent.exp.canceled = true
	}
	c.ll.Remove(ele)
	delete(c.cache, ent.key)
}

func (c *Cache[K, V]) expirationProcessor() {
	var timer *time.Timer

	for {
		c.mu.Lock()
		var waitDuration time.Duration
		now := time.Now()
		if c.expHeap.Len() == 0 {
			// No items with TTL. Wait for a long time (or until an update).
			waitDuration = time.Hour
		} else {
			// Peek at the top of the heap.
			next := c.expHeap[0]
			// If the entry is canceled, remove it immediately.
			if next.canceled {
				heap.Pop(&c.expHeap)
				c.mu.Unlock()
				continue
			}
			if now.Before(next.expiration) {
				waitDuration = next.expiration.Sub(now)
			} else {
				// Expired – set waitDuration to 0.
				waitDuration = 0
			}
		}
		c.mu.Unlock()

		// Create or reset the timer.
		if timer == nil {
			timer = time.NewTimer(waitDuration)
		} else {
			if !timer.Stop() {
				// Drain the channel if needed.
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(waitDuration)
		}

		// Wait for the timer to fire, an update, or shutdown.
		select {
		case <-timer.C:
			// Time to remove expired items.
		case <-c.updateCh:
			// An update was signaled; loop around to recalc waitDuration.
			continue
		case <-c.done:
			timer.Stop()
			return
		}

		// Remove all expired entries.
		c.mu.Lock()
		now = time.Now()
		for c.expHeap.Len() > 0 {
			next := c.expHeap[0]
			// Skip canceled entries.
			if next.canceled {
				heap.Pop(&c.expHeap)
				continue
			}
			if now.Before(next.expiration) {
				break
			}
			// Pop from the heap.
			heap.Pop(&c.expHeap)
			// Remove from cache if it still exists and its expiration matches.
			if ele, ok := c.cache[next.key]; ok {
				ent := ele.Value.(*entry[K, V])
				// Only remove if the stored expiration is expired.
				if !ent.expiration.IsZero() && !now.Before(ent.expiration) {
					c.ll.Remove(ele)
					delete(c.cache, next.key)
				}
			}
		}
		c.mu.Unlock()
	}
}

// Removes a key from the cache if it exists.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.cache[key]; ok {
		ent := ele.Value.(*entry[K, V])
		if ent.exp != nil {
			ent.exp.canceled = true
		}
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
	// Reset the expiration heap.
	c.expHeap = nil
	heap.Init(&c.expHeap)
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
		c.removeOldestLocked()
	}
}

// Stops the background expiration goroutine.
func (c *Cache[K, V]) Close() {
	close(c.done)
}
