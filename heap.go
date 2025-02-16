package goutte

import "time"

// Entry for the expiration heap.
type expEntry[K comparable] struct {
	key        K
	expiration time.Time
	index      int  // needed by heap.Interface for update/removal
	canceled   bool // indicates that this entry is outdated/canceled
}

// expHeap is a min-heap of *expEntry items.
type expHeap[K comparable] []*expEntry[K]

func (h expHeap[K]) Len() int { return len(h) }

func (h expHeap[K]) Less(i, j int) bool {
	return h[i].expiration.Before(h[j].expiration)
}

func (h expHeap[K]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *expHeap[K]) Push(x interface{}) {
	n := len(*h)
	item := x.(*expEntry[K])
	item.index = n
	*h = append(*h, item)
}

func (h *expHeap[K]) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*h = old[0 : n-1]
	return item
}
