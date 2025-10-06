package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndpointToSubsystem(t *testing.T) {
	assert := assert.New(t)

	a := NewAPIMetrics("test", map[string]string{})

	endpoint := "foobar"
	subsystem := a.endpointToSubsystem(endpoint)
	assert.Equal("foobar", subsystem)

	endpoint = "foo.bar"
	subsystem = a.endpointToSubsystem(endpoint)
	assert.Equal("foo_bar", subsystem)

	endpoint = "foo/bar"
	subsystem = a.endpointToSubsystem(endpoint)
	assert.Equal("foo_bar", subsystem)

	endpoint = "foo/bar/"
	subsystem = a.endpointToSubsystem(endpoint)
	assert.Equal("foo_bar", subsystem)

	endpoint = "http://foo.com/bar"
	subsystem = a.endpointToSubsystem(endpoint)
	assert.Equal("foo_com", subsystem)

	endpoint = "//foo.com/bar"
	subsystem = a.endpointToSubsystem(endpoint)
	assert.Equal("foo_com", subsystem)

	endpoint = "foo.com/bar"
	subsystem = a.endpointToSubsystem(endpoint)
	assert.Equal("foo_com_bar", subsystem)
}
