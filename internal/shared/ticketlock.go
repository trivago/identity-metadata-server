package shared

import (
	"container/heap"
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type TicketLock struct {
	nextTicket      uint64
	activeTicket    uint64
	pause           *time.Ticker
	canceledTickets *HeapUint64
	ticketGuard     *sync.Mutex
}

// NewTicketLock creates a new ticket lock with the given granularity.
// The granularity is the time to wait between each lock acquisition check.
// The granularity should be small enough to not block the main thread for too
// long, but large enough to not waste too much time.
// A granularity of 5-10 milliseconds is a good starting point.
func NewTicketLock(granularity time.Duration) *TicketLock {
	return &TicketLock{
		nextTicket:      1,
		activeTicket:    1,
		pause:           time.NewTicker(granularity),
		canceledTickets: &HeapUint64{},
		ticketGuard:     &sync.Mutex{},
	}
}

// Close stops the ticket lock.
// It is important to call this method to avoid memory leaks.
func (l *TicketLock) Close() {
	l.pause.Stop()
}

// IsLocked returns true if the lock is currently held by a thread.
// Please note that this status can change right after the call to IsLocked().
// I.e. this is not a reliable way to check if the lock is currently held by a
// thread. It is only meant for debugging purposes.
func (l *TicketLock) IsLocked() bool {
	return atomic.LoadUint64(&l.activeTicket) != atomic.LoadUint64(&l.nextTicket)
}

// Lock tries to acquire a lock in a FIFO way.
func (l *TicketLock) Lock() uint64 {
	return l.LockWithContext(context.Background())
}

// LockWithContext tries to acquire a lock in a FIFO way.
// It returns 0 when the lock failed to be acquired due to a context
// cancellation or a timeout.
// If the lock was acquired, it returns the ticket number of the lock.
func (l *TicketLock) LockWithContext(ctx context.Context) uint64 {
	ticket := atomic.AddUint64(&l.nextTicket, 1) - 1

	for {
		if atomic.LoadUint64(&l.activeTicket) == ticket {
			return ticket
		}

		select {
		// The first tick may come early, as we keep the ticker running even if
		// no lock is acquired. This is fine, as we need this to a) check on the
		// context and b) instruct the processor to yield the CPU during waiting.
		case <-l.pause.C:
			continue

		case <-ctx.Done():
			// We need to keep track of canceled tickets as tickets are linearly
			// ordered. If we don't do this, we cannot properly unlock the lock
			// in the correct order.
			l.ticketGuard.Lock()
			defer l.ticketGuard.Unlock()
			heap.Push(l.canceledTickets, ticket)
			return 0
		}
	}
}

// Unlock releases the lock.
func (l *TicketLock) Unlock() {
	l.ticketGuard.Lock()
	defer l.ticketGuard.Unlock()

	for {
		ticket := atomic.AddUint64(&l.activeTicket, 1)
		lastCanceledTicket, hasCanceledTickets := l.canceledTickets.Peek()

		switch {
		// No canceled tickets, we can return
		case !hasCanceledTickets:
			return

		// The last canceled ticket is the same as the current ticket.
		// We need to try again with the next ticket (which might also be
		// canceled).
		case lastCanceledTicket == ticket:
			heap.Pop(l.canceledTickets)

		// There are canceled tickets, but the current ticket is smaller than
		// the first canceled ticket.
		default:
			return
		}
	}
}
