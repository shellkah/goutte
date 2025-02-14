# Goutte LRU Cache

A generic, thread-safe LRU (Least Recently Used) cache implemented in Go using generics (Go 1.18+). This cache provides fast point queries via a hash map and maintains access order with a doubly linked list, automatically evicting the least recently used item when the cache exceeds its capacity.

## Features

- **Generics**: Specify key and value types at creation time for compile-time type safety.
- **Thread-Safe**: Safe for concurrent access using a mutex.
- **LRU Eviction Policy**: Automatically removes the least recently used entry when adding new items beyond the specified capacity.
- **Fast Lookups**: Uses a hash map for O(1) average-time complexity for queries.
- **Simple API**: Provides basic operations such as `Get`, `Put`, and `Delete`.

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

	"github.com/your_username/your_repository/lru" // update with your actual module path
)

func main() {
	// Create a cache where keys are strings and values are ints.
	cache := lru.New[string, int](3)

	// Insert key-value pairs.
	cache.Put("a", 1)
	cache.Put("b", 2)
	cache.Put("c", 3)

	// Retrieve a value.
	if val, found := cache.Get("a"); found {
		fmt.Println("Value for 'a':", val)
	} else {
		log.Println("Key 'a' not found")
	}

	// Adding a new key causes eviction of the least recently used item.
	cache.Put("d", 4)

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