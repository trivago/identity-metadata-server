package shared

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTicketLock(t *testing.T) {
	assert := assert.New(t)

	lock := NewTicketLock(time.Millisecond)

	ticket1 := lock.Lock()
	assert.NotZero(ticket1, "Lock should return a non-zero ticket")
	assert.Equal(uint64(1), ticket1, "Lock should return the first ticket")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ticket2 := lock.LockWithContext(ctx)
	assert.Zero(ticket2, "LockWithContext should return a zero ticket as it timed out before the lock was acquired")

	lock.Unlock()

	ticket3 := lock.Lock()
	assert.NotZero(ticket3, "Lock should return a non-zero ticket after the previous lock was released")
	assert.NotEqual(ticket1, ticket3, "Lock should return a different ticket after the previous lock was released")
	assert.Equal(uint64(3), ticket3, "Lock should return the third ticket, as the second lock was aborted")
}
