package goutte

import (
	"sync"
	"testing"
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
