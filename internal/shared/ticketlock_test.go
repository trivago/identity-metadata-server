package shared

import (
	"context"
	"slices"
	"sync"
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
	assert.True(lock.IsLocked(), "Lock should return a non-zero ticket")

	// Test timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ticket2 := lock.LockWithContext(ctx)
	assert.Zero(ticket2, "LockWithContext should return a zero ticket as it timed out before the lock was acquired")

	// Test release
	lock.Unlock()
	assert.False(lock.IsLocked(), "Lock should return a zero ticket after the lock was released")

	// Test if release properly increments the active ticket
	ticket3 := lock.Lock()
	assert.NotZero(ticket3, "Lock should return a non-zero ticket after the previous lock was released")
	assert.NotEqual(ticket1, ticket3, "Lock should return a different ticket after the previous lock was released")
	assert.Equal(uint64(3), ticket3, "Lock should return the third ticket, as the second lock was aborted")

	// Test if release properly increments the active ticket with consecutive discards
	ticket4 := lock.LockWithContext(ctx)
	assert.Zero(ticket4, "LockWithContext should return a zero ticket as it timed out before the lock was acquired")
	ticket5 := lock.LockWithContext(ctx)
	assert.Zero(ticket5, "LockWithContext should return a zero ticket as it timed out before the lock was acquired")

	lock.Unlock()
	ticket6 := lock.Lock()
	assert.NotZero(ticket6, "Lock should return a non-zero ticket after the previous lock was released")
	assert.NotEqual(ticket3, ticket6, "Lock should return a different ticket after the previous lock was released")
	assert.Equal(uint64(6), ticket6, "Lock should return the sixth ticket, as the fourth and fifth locks were aborted")
}

func TestTicketLockConcurrency(t *testing.T) {
	assert := assert.New(t)

	lock := NewTicketLock(time.Millisecond)

	wg := sync.WaitGroup{}
	wg.Add(10)
	done := make(chan struct{})
	order := make([]uint64, 0, 10)

	go func() {
		for i := 0; i < 10; i++ {
			go func() {
				defer wg.Done()
				ticket := lock.Lock()
				order = append(order, ticket)
				time.Sleep(2 * time.Millisecond)
				lock.Unlock()
			}()
		}
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		break
	case <-time.After(time.Second):
		t.Fatal("Test should have finished within 1 second")
	}

	assert.True(slices.IsSorted(order), "Locks should be ordered")
	assert.Equal(uint64(10), order[len(order)-1], "Last ticket should be equal to the number of runs")
	assert.False(lock.IsLocked(), "Lock should not be locked after all locks have been released")
}

func TestTicketLockConcurrencyWithContext(t *testing.T) {
	assert := assert.New(t)

	lock := NewTicketLock(time.Millisecond)

	wg := sync.WaitGroup{}
	wg.Add(10)
	done := make(chan struct{})
	order := make([]uint64, 0, 10)

	go func() {
		for i := 0; i < 10; i++ {
			go func() {
				defer wg.Done()

				for {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
					defer cancel()

					ticket := lock.LockWithContext(ctx)
					time.Sleep(5 * time.Millisecond)
					if ticket != 0 {
						order = append(order, ticket)
						lock.Unlock()
						return
					}
				}
			}()
		}
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		break
	case <-time.After(10 * time.Second):
		t.Fatal("Test should have finished within 10 seconds")
	}

	assert.True(slices.IsSorted(order), "Locks should be ordered")
	assert.Greater(order[len(order)-1], uint64(10), "Last ticket should be greater than 10 as locks should be aborted if the context is done")
	assert.False(lock.IsLocked(), "Lock should not be locked after all locks have been released")
}

func TestTicketLockConcurrencyWithContextAndPause(t *testing.T) {
	assert := assert.New(t)

	wg := sync.WaitGroup{}
	wg.Add(1)
	done := make(chan struct{})

	// Granularity must be larger than the timeout of the context to test the
	// behavior.
	lock := NewTicketLock(100 * time.Millisecond)

	ticket1 := lock.Lock()
	assert.NotZero(ticket1, "Lock should return a non-zero ticket")

	// Unlock while the second lock is waiting for the pause.
	// As granularity is larger than this delay the second lock is still waiting
	// when the unlock is called.
	time.AfterFunc(5*time.Millisecond, func() {
		lock.Unlock()
	})

	// Make sure to have a context timeout that is between the granularity and
	// the delay of the unlock.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// If the test case fails, we have a deadlock here
	go func() {
		defer wg.Done()
		ticket2 := lock.LockWithContext(ctx)
		assert.NotZero(ticket2, "LockWithContext should return a non-zero ticket as the lock was acquired before the context was done")
		lock.Unlock()
		close(done)
	}()

	select {
	case <-done:
		break
	case <-time.After(1 * time.Second):
		t.Fatal("Test should have finished within 1 seconds")
	}
}
