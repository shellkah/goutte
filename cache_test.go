package goutte

import (
	"sync"
	"testing"
	"time"
)

func TestCacheBasic(t *testing.T) {
	cache := NewCache[string, int](2)
	defer cache.Close()
	cache.Set("a", 1)

	if val, ok := cache.Get("a"); !ok || val != 1 {
		t.Errorf("Expected key 'a' to have value 1, got %v (found: %v)", val, ok)
	}
}

func TestCacheEviction(t *testing.T) {
	cache := NewCache[string, int](2)
	defer cache.Close()
	cache.Set("a", 1)
	cache.Set("b", 2)

	// Access "a" so "b" becomes the least recently used.
	if _, ok := cache.Get("a"); !ok {
		t.Error("Expected key 'a' to be present")
	}

	// Adding a new item should evict "b".
	cache.Set("c", 3)
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
	cache := NewCache[string, int](2)
	defer cache.Close()
	cache.Set("a", 1)
	cache.Set("a", 10)

	if val, ok := cache.Get("a"); !ok || val != 10 {
		t.Errorf("Expected key 'a' to have updated value 10, got %v (found: %v)", val, ok)
	}
}

func TestCacheDelete(t *testing.T) {
	cache := NewCache[string, int](2)
	defer cache.Close()
	cache.Set("a", 1)
	cache.Delete("a")

	if _, ok := cache.Get("a"); ok {
		t.Error("Expected key 'a' to be deleted")
	}
}

func TestCacheConcurrency(t *testing.T) {
	cache := NewCache[int, int](1000)
	defer cache.Close()
	var wg sync.WaitGroup

	// Insert values concurrently.
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cache.Set(i, i*10)
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
	cache := NewCache[string, int](2)
	defer cache.Close()
	cache.Set("a", 1)
	cache.Set("b", 2)

	cache.Dump()

	if _, ok := cache.Get("a"); ok {
		t.Error("Expected cache to be empty after Dump, but found key 'a'")
	}
	if _, ok := cache.Get("b"); ok {
		t.Error("Expected cache to be empty after Dump, but found key 'b'")
	}

	cache.Set("c", 3)
	if val, ok := cache.Get("c"); !ok || val != 3 {
		t.Errorf("Expected key 'c' to have value 3, got %v (found: %v)", val, ok)
	}
}

func TestCacheTTL(t *testing.T) {
	cache := NewCache[string, int](2)
	defer cache.Close()

	// Insert an item with a TTL of 50 milliseconds.
	cache.SetWithTTL("a", 1, 50*time.Millisecond)

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
	cache := NewCache[string, int](3)
	defer cache.Close()
	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)

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
	cache.Set("a", 10)
	cache.Set("d", 4)
	cache.Set("e", 5)
	cache.Set("f", 6)

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

func TestCacheTTLUpdate(t *testing.T) {
	cache := NewCache[string, int](2)
	defer cache.Close()

	// Set the key "update" with a TTL of 50ms.
	cache.SetWithTTL("update", 1, 50*time.Millisecond)

	// Wait for 40ms (still within the initial TTL).
	time.Sleep(40 * time.Millisecond)

	// Update the same key with a new TTL of 100ms from now.
	cache.SetWithTTL("update", 1, 100*time.Millisecond)

	// Wait another 20ms. The original 50ms TTL would have expired by now,
	// but since we updated it, the key should still be present.
	time.Sleep(20 * time.Millisecond)
	if val, ok := cache.Get("update"); !ok || val != 1 {
		t.Errorf("Expected key 'update' to exist after TTL update, got %v (found: %v)", val, ok)
	}

	// Wait for a period that exceeds the new TTL.
	time.Sleep(90 * time.Millisecond)
	if _, ok := cache.Get("update"); ok {
		t.Error("Expected key 'update' to have expired after updated TTL, but it was found")
	}
}

func TestCacheTTLCancel(t *testing.T) {
	cache := NewCache[string, int](2)
	defer cache.Close()

	cache.SetWithTTL("cancel", 1, 50*time.Millisecond)

	// A TTL of 0 is intended to cancel any existing expiration.
	cache.SetWithTTL("cancel", 1, 0)

	// Wait for longer than the original TTL.
	time.Sleep(70 * time.Millisecond)

	// The key should still exist because the expiration was canceled.
	if val, ok := cache.Get("cancel"); !ok || val != 1 {
		t.Errorf("Expected key 'cancel' to remain after TTL cancellation, got %v (found: %v)", val, ok)
	}
}
