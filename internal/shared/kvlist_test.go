package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKVList(t *testing.T) {
	assert := assert.New(t)

	kv := NewKVList[string, string]()
	assert.NotNil(kv, "NewKVList should return a non-nil KVList")

	// Test adding key-value pairs
	kv.Add("key1", "value1")
	kv.Add("key2", "value2")
	assert.Equal(2, kv.Len(), "KVList should have 2 items after adding 2 pairs")

	// Test getting values
	value1, ok1 := kv.Get("key1")
	assert.True(ok1, "Get should return true for existing key")
	assert.Equal("value1", value1, "Get should return the correct value for key1")

	value2, ok2 := kv.Get("key2")
	assert.True(ok2, "Get should return true for existing key")
	assert.Equal("value2", value2, "Get should return the correct value for key2")

	// Test getting a non-existing key
	value3, ok3 := kv.Get("key3")
	assert.False(ok3, "Get should return false for non-existing key")
	assert.Empty(value3, "Get should return an empty string for non-existing key")

	// Test removing a key
	kv.Remove("key1")
	assert.Equal(1, kv.Len(), "KVList should have 1 item after removing key1")

	// Test that removed key cannot be retrieved
	value4, ok4 := kv.Get("key1")
	assert.False(ok4, "Get should return false for removed key")
	assert.Empty(value4, "Get should return an empty string for removed key")

	// Test clearing the list
	kv.Clear()
	assert.Equal(0, kv.Len(), "KVList should be empty after clearing")
}

func TestKVListOrder(t *testing.T) {
	assert := assert.New(t)

	kv := NewKVList[string, string]()
	kv.Add("key1", "value1")
	kv.Add("key2", "value2")
	kv.Add("key3", "value3")

	// Check the order of keys via ForEach
	var keys []string
	kv.ForEach(func(key string, value string) bool {
		keys = append(keys, key)
		return true // continue iterating
	})

	assert.Equal([]string{"key1", "key2", "key3"}, keys, "Keys should be in the order they were added")

	// Check that values are correct
	expectedValues := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	kv.ForEach(func(key string, value string) bool {
		if expectedValue, exists := expectedValues[key]; exists {
			assert.Equal(expectedValue, value, "Value for %s should be %s", key, expectedValue)
		}
		return true // continue iterating
	})

	// Test removing a key and checking order
	kv.Remove("key2")

	// Check that value for removed key is no longer accessible
	value2, ok2 := kv.Get("key2")
	assert.False(ok2, "Get should return false for removed key2")
	assert.Empty(value2, "Get should return an empty string for removed key2")

	keys = nil // reset keys slice
	kv.ForEach(func(key string, value string) bool {
		keys = append(keys, key)
		return true // continue iterating
	})

	assert.Equal([]string{"key1", "key3"}, keys, "Keys should be in the order after removing key2")

	// Test adding a new key after removal
	kv.Add("key4", "value4")
	assert.Equal(3, kv.Len(), "KVList should have 3 items after adding key4")
	keys = nil // reset keys slice
	kv.ForEach(func(key string, value string) bool {
		keys = append(keys, key)
		return true // continue iterating
	})

	assert.Equal([]string{"key1", "key3", "key4"}, keys, "Keys should be in the order after adding key4")
}
