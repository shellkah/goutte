package goutte_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/shellkah/goutte"
)

// Simulates a heavy concurrent workload against the cache.
func BenchmarkCacheLoad(b *testing.B) {
	cacheCapacity := 10000
	c := goutte.NewCache[string, int](cacheCapacity)

	numPrepopulate := 5000
	for i := 0; i < numPrepopulate; i++ {
		key := "key" + strconv.Itoa(i)
		c.Set(key, i)
	}

	var wg sync.WaitGroup

	numWorkers := 100
	opsPerWorker := b.N / numWorkers
	if opsPerWorker < 1 {
		opsPerWorker = 1
	}

	b.ResetTimer()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			// Each goroutine has its own random generator.
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))
			for j := 0; j < opsPerWorker; j++ {
				op := r.Intn(100)
				// Use a key from a wider range to force some misses.
				keyID := r.Intn(15000)
				key := fmt.Sprintf("key%d", keyID)
				switch {
				case op < 50:
					// 50% chance: read (Get) operation.
					_, _ = c.Get(key)
				case op < 80:
					// 30% chance: write (Set) operation without TTL.
					c.Set(key, r.Intn(1000000))
				case op < 90:
					// 10% chance: write (SetWithTTL) with a short TTL (simulate expiry).
					ttl := time.Duration(r.Intn(100)) * time.Millisecond
					c.SetWithTTL(key, r.Intn(1000000), ttl)
				default:
					// 10% chance: Delete the key.
					c.Delete(key)
				}
			}
		}(i)
	}

	// Periodically adjusts the capacity.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			time.Sleep(50 * time.Millisecond)
			newCap := 5000 + rand.Intn(10000)
			c.SetCapacity(newCap)
		}
	}()

	wg.Wait()

	c.Close()
}

func TestCacheLoad(t *testing.T) {
	const numOperations = 100000

	cacheCapacity := 10000
	c := goutte.NewCache[string, int](cacheCapacity)

	for i := 0; i < 5000; i++ {
		key := fmt.Sprintf("key%d", i)
		c.Set(key, i)
	}

	var wg sync.WaitGroup
	numWorkers := 50
	opsPerWorker := numOperations / numWorkers
	if opsPerWorker < 1 {
		opsPerWorker = 1
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))
			for j := 0; j < opsPerWorker; j++ {
				op := r.Intn(100)
				keyID := r.Intn(15000)
				key := "key" + strconv.Itoa(keyID)
				switch {
				case op < 50:
					// Get operation.
					_, _ = c.Get(key)
				case op < 80:
					// Set without TTL.
					c.Set(key, r.Intn(1000000))
				case op < 90:
					// Set with TTL.
					ttl := time.Duration(r.Intn(100)) * time.Millisecond
					c.SetWithTTL(key, r.Intn(1000000), ttl)
				default:
					// Delete operation.
					c.Delete(key)
				}
			}
		}(i)
	}
	wg.Wait()
	c.Close()
}
