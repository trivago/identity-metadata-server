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
	granularity     time.Duration
	canceledTickets *HeapUint64
	ticketGuard     *sync.Mutex
}

func NewTicketLock(granularity time.Duration) *TicketLock {
	return &TicketLock{
		nextTicket:      1,
		activeTicket:    1,
		granularity:     granularity,
		canceledTickets: &HeapUint64{},
		ticketGuard:     &sync.Mutex{},
	}
}

// Lock tries to aquire a lock in a FIFO way.
func (l *TicketLock) Lock() uint64 {
	return l.LockWithContext(context.Background())
}

// LockWithContext tries to aquire a lock in a FIFO way.
// It returns true if the lock was acquired, false if the context was cancelled
// while waiting for the lock.
func (l *TicketLock) LockWithContext(ctx context.Context) uint64 {
	ticket := atomic.AddUint64(&l.nextTicket, 1) - 1

	for {
		if atomic.LoadUint64(&l.activeTicket) == ticket {
			return ticket
		}

		select {
		case <-time.After(l.granularity):
			continue

		case <-ctx.Done():
			l.Abort(ticket)
			return 0
		}
	}
}

// Abort marks a ticket as canceled.
func (l *TicketLock) Abort(ticket uint64) {
	l.ticketGuard.Lock()
	defer l.ticketGuard.Unlock()

	heap.Push(l.canceledTickets, ticket)
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
