package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEqualUnordered(t *testing.T) {
	assert := assert.New(t)

	a := []string{"a", "b", "c"}
	b := []string{"c", "b", "a"}
	c := []string{"a", "b", "d"}

	assert.True(EqualUnordered(a, a))
	assert.True(EqualUnordered(a, b))
	assert.False(EqualUnordered(a, c))

	assert.False(EqualUnordered(a, []string{"a", "b"}))
	assert.False(EqualUnordered(a, []string{"a", "b", "c", "d"}))
	assert.False(EqualUnordered(a, []string{}))
	assert.True(EqualUnordered([]string{}, []string{}), "Expected empty slices to be equal")
}

func TestEqualUnorderedFunc(t *testing.T) {
	assert := assert.New(t)

	a := []struct {
		name string
		age  int
	}{
		{"Alice", 30},
		{"Bob", 25},
	}

	b := []struct {
		name string
		age  int
	}{
		{"Bob", 25},
		{"Alice", 30},
	}

	c := []struct {
		name string
		age  int
	}{
		{"Alice", 30},
	}

	assert.True(EqualUnorderedFunc(a, a, func(x, y struct {
		name string
		age  int
	}) bool {
		return x.name == y.name && x.age == y.age
	}))

	assert.True(EqualUnorderedFunc(a, b, func(x, y struct {
		name string
		age  int
	}) bool {
		return x.name == y.name && x.age == y.age
	}))

	assert.False(EqualUnorderedFunc(a, c, func(x, y struct {
		name string
		age  int
	}) bool {
		return x.name == y.name && x.age == y.age
	}))
}
