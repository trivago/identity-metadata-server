package shared

// HeapUint64 implements a min-heap of uint64 values on top of a slice.
// Use this type with the container/heap package.
type HeapUint64 []uint64

// Len returns the number of elements in the heap.
func (h HeapUint64) Len() int {
	return len(h)
}

// Less returns true if the element at index i is less than the element at index j.
func (h HeapUint64) Less(i, j int) bool { return h[i] < h[j] }

// Swap swaps the elements at index i and j.
func (h HeapUint64) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

// Push adds a new element to the heap.
// Use heap.Push(h, x) instead of this function.
func (h *HeapUint64) Push(x any) {
	*h = append(*h, x.(uint64))
}

// Peek returns the smallest element from the heap without removing it.
// It returns false if the heap is empty.
func (h *HeapUint64) Peek() (uint64, bool) {
	if len(*h) == 0 {
		return 0, false
	}
	return (*h)[0], true
}

// Pop removes and returns the smallest element from the heap.
// Use heap.Pop(h) instead of this function.
func (h *HeapUint64) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
