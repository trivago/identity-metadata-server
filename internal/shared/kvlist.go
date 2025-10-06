package shared

import "slices"

// KVList is a key-value list that maintains the order of keys.
// This datatype is not thread-safe and should only be used in
// a single-threaded context or with external synchronization.
type KVList[K comparable, V any] struct {
	keys   []K
	values map[K]V
}

// NewKVList creates a new KVList instance.
func NewKVList[K comparable, V any]() *KVList[K, V] {
	return &KVList[K, V]{
		keys:   make([]K, 0),
		values: make(map[K]V),
	}
}

// Clear removes all keys and values from the KVList.
func (l *KVList[K, V]) Clear() {
	l.keys = make([]K, 0)
	l.values = make(map[K]V)
}

// Add adds or updates a key-value pair in the KVList.
// If the key does not exist, it is added to the end of the keys slice.
// If the key already exists, its value is updated.
func (l *KVList[K, V]) Add(key K, value V) {
	if _, exists := l.values[key]; !exists {
		l.keys = append(l.keys, key)
	}
	l.values[key] = value
}

// Remove removes a key and its associated value from the KVList.
// This operation is O(n) in the worst case because we need to preserve the
// order of keys.
// If the key does not exist, it does nothing.
func (l *KVList[K, V]) Remove(key K) {
	if _, exists := l.values[key]; exists {
		delete(l.values, key)
		if i := slices.Index(l.keys, key); i >= 0 {
			l.keys = slices.Delete(l.keys, i, i+1)
		}
	}
}

// Get retrieves the value associated with a key.
// It returns the value and a boolean indicating whether the key exists.
// If the key does not exist, it returns the zero value of V and false.
func (l *KVList[K, V]) Get(key K) (V, bool) {
	if value, exists := l.values[key]; exists {
		return value, true
	}
	var zeroValue V
	return zeroValue, false
}

// ForEach iterates over each key-value pair in the KVList.
// It calls the provided function with each key and its associated value.
// If the called function returns false, the iteration stops.
func (l *KVList[K, V]) ForEach(fn func(key K, value V) bool) {
	for _, key := range l.keys {
		if value, exists := l.values[key]; exists {
			if !fn(key, value) {
				break
			}
		}
	}
}

// Len returns the number of key-value pairs in the KVList.
func (l *KVList[K, V]) Len() int {
	return len(l.keys)
}
