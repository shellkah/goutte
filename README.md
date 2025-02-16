# Goutte Cache

Thread-safe and type-safe LRU (Least Recently Used) cache implemented in Go. This cache provides fast point queries via a hash map, use heap-based expiration to handle TTLs and maintains access order with a doubly linked list, automatically evicting the least recently used item when the cache exceeds its capacity.

## Features

- **Generics**: Specify key and value types at creation time for compile-time type safety.
- **Thread-Safe**: Safe for concurrent access using a mutex.
- **LRU Eviction Policy**: Automatically removes the least recently used entry when adding new items beyond the specified capacity.
- **Optional TTL**: Automatically removes expired items with precision with a min-heap (priority queue) to keep track of expiration times.
- **Fast Lookups**: Uses a hash map for O(1) average-time complexity for queries.
- **Simple API**: Provides basic operations such as `Get`, `Set`, and `Delete`.

## Installation

To include the cache in your project, use Go modules. In your project directory, run:

```bash
go get github.com/shellkah/goutte
```

## Usage

Here's an example of how to use the cache:

```go
package main

import (
	"fmt"
	"log"

	"github.com/shellkah/goutte"
)

func main() {
	// Create a cache where keys are strings and values are ints.
	cache := goutte.NewCache[string, int](3)
	defer cache.Close()

	// Insert key-value pairs.
	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)

	// Retrieve a value.
	if val, found := cache.Get("a"); found {
		fmt.Println("Value for 'a':", val)
	} else {
		log.Println("Key 'a' not found")
	}

	// Adding a new key causes eviction of the least recently used item.
	cache.Set("d", 4)

	// 'b' should be evicted if it was the least recently used.
	if _, found := cache.Get("b"); !found {
		fmt.Println("Key 'b' was evicted.")
	}
}
```

## Contributing

Contributions are welcome! Please open issues or submit pull requests if you have any ideas, bug fixes, or enhancements.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
