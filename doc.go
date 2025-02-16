// Package goutte provides a generic, thread-safe LRU (Least Recently Used) cache implementation in Go.
//
// The cache is implemented using a hash map for fast point queries and a doubly linked list (via the
// standard library’s container/list package) to track the usage order. When the cache reaches its
// capacity, the least recently used item is evicted automatically.
//
// The cache is generic, allowing you to specify the types for keys and values at creation time (using
// Go 1.18+ generics). Keys must be comparable and values can be of any type. All operations (Get, Set,
// Delete) are safe for concurrent use.
//
// ## Usage Example
//
//	package main
//
//	import (
//		"fmt"
//		"github.com/shellkah/goutte"
//	)
//
//	func main() {
//		// Create a cache where keys are strings and values are ints.
//		cache := goutte.NewCache[string, int](3)
//
//		// Insert key-value pairs.
//		cache.Set("a", 1)
//		cache.Set("b", 2)
//		cache.Set("c", 3)
//
//		// Retrieve a value.
//		if val, found := cache.Get("a"); found {
//			fmt.Println("Value for 'a':", val)
//		}
//
//		// Inserting a new value when the cache is full causes eviction of the least recently used item.
//		cache.Set("d", 4)
//		if _, found := cache.Get("b"); !found {
//			fmt.Println("Key 'b' was evicted.")
//		}
//	}
package goutte
