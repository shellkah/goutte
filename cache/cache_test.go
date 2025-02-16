package goutte

import (
	"sync"
	"testing"
	"time"
)

func TestCacheBasic(t *testing.T) {
	cache := New[string, int](2)
	cache.Put("a", 1)

	if val, ok := cache.Get("a"); !ok || val != 1 {
		t.Errorf("Expected key 'a' to have value 1, got %v (found: %v)", val, ok)
	}
}

func TestCacheEviction(t *testing.T) {
	cache := New[string, int](2)
	cache.Put("a", 1)
	cache.Put("b", 2)

	// Access "a" so "b" becomes the least recently used.
	if _, ok := cache.Get("a"); !ok {
		t.Error("Expected key 'a' to be present")
	}

	// Adding a new item should evict "b".
	cache.Put("c", 3)
	if _, ok := cache.Get("b"); ok {
		t.Error("Expected key 'b' to be evicted")
	}

	// "a" and "c" should be available.
	if val, ok := cache.Get("a"); !ok || val != 1 {
		t.Errorf("Expected key 'a' to have value 1, got %v (found: %v)", val, ok)
	}
	if val, ok := cache.Get("c"); !ok || val != 3 {
		t.Errorf("Expected key 'c' to have value 3, got %v (found: %v)", val, ok)
	}
}

func TestCacheUpdate(t *testing.T) {
	cache := New[string, int](2)
	cache.Put("a", 1)
	cache.Put("a", 10)

	if val, ok := cache.Get("a"); !ok || val != 10 {
		t.Errorf("Expected key 'a' to have updated value 10, got %v (found: %v)", val, ok)
	}
}

func TestCacheDelete(t *testing.T) {
	cache := New[string, int](2)
	cache.Put("a", 1)
	cache.Delete("a")

	if _, ok := cache.Get("a"); ok {
		t.Error("Expected key 'a' to be deleted")
	}
}

func TestCacheConcurrency(t *testing.T) {
	cache := New[int, int](1000)
	var wg sync.WaitGroup

	// Insert values concurrently.
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cache.Put(i, i*10)
		}(i)
	}
	wg.Wait()

	// Retrieve values concurrently.
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if val, ok := cache.Get(i); ok && val != i*10 {
				t.Errorf("For key %d, expected %d but got %d", i, i*10, val)
			}
		}(i)
	}
	wg.Wait()
}

func TestCacheDump(t *testing.T) {
	cache := New[string, int](2)
	cache.Put("a", 1)
	cache.Put("b", 2)

	cache.Dump()

	if _, ok := cache.Get("a"); ok {
		t.Error("Expected cache to be empty after Dump, but found key 'a'")
	}
	if _, ok := cache.Get("b"); ok {
		t.Error("Expected cache to be empty after Dump, but found key 'b'")
	}

	cache.Put("c", 3)
	if val, ok := cache.Get("c"); !ok || val != 3 {
		t.Errorf("Expected key 'c' to have value 3, got %v (found: %v)", val, ok)
	}
}

func TestCacheTTL(t *testing.T) {
	cache := New[string, int](2)

	// Insert an item with a TTL of 50 milliseconds.
	cache.PutWithTTL("a", 1, 50*time.Millisecond)

	// Immediately retrieving should succeed.
	if val, ok := cache.Get("a"); !ok || val != 1 {
		t.Errorf("Expected key 'a' to have value 1, got %v (found: %v)", val, ok)
	}

	// Wait for the TTL to expire.
	time.Sleep(100 * time.Millisecond)

	// Now the item should have expired.
	if _, ok := cache.Get("a"); ok {
		t.Error("Expected key 'a' to have expired, but it was found")
	}
}

func TestCacheSetCapacity(t *testing.T) {
	// Start with a capacity of 3.
	cache := New[string, int](3)
	cache.Put("a", 1)
	cache.Put("b", 2)
	cache.Put("c", 3)

	// Reduce capacity to 2. This should evict the least recently used item.
	cache.SetCapacity(2)

	// Count the number of items present.
	count := 0
	if _, ok := cache.Get("a"); ok {
		count++
	}
	if _, ok := cache.Get("b"); ok {
		count++
	}
	if _, ok := cache.Get("c"); ok {
		count++
	}
	if count != 2 {
		t.Errorf("Expected 2 items after reducing capacity, got %d", count)
	}

	// Increase capacity to 5.
	cache.SetCapacity(5)
	cache.Put("a", 10)
	cache.Put("d", 4)
	cache.Put("e", 5)
	cache.Put("f", 6)

	// Now expect 5 items in total.
	count = 0
	if _, ok := cache.Get("a"); ok {
		count++
	}
	if _, ok := cache.Get("b"); ok {
		count++
	}
	if _, ok := cache.Get("c"); ok {
		count++
	}
	if _, ok := cache.Get("d"); ok {
		count++
	}
	if _, ok := cache.Get("e"); ok {
		count++
	}
	if _, ok := cache.Get("f"); ok {
		count++
	}
	if count != 5 {
		t.Errorf("Expected 5 items after increasing capacity and adding new items, got %d", count)
	}
}
