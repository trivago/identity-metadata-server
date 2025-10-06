package shared

import (
	"slices"
)

// EqualUnordered compares two slices for equality
// without considering the order of elements.
func EqualUnordered[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for _, v := range a {
		if !slices.Contains(b, v) {
			return false
		}
	}
	return true
}

// EqualUnorderedFunc compares two slices for equality
// without considering the order of elements, using a custom equality function.
func EqualUnorderedFunc[T any](a, b []T, equal func(T, T) bool) bool {
	if len(a) != len(b) {
		return false
	}

outerLoop:
	for _, va := range a {
		for _, vb := range b {
			if equal(va, vb) {
				continue outerLoop
			}
		}
		return false
	}
	return true
}
