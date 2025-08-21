package emit

import (
	"sync"
	"sync/atomic"

	"github.com/kelindar/event"
)

const segmentSize = 64

// segment represents a fixed-size buffer that holds events
type segment[T any] struct {
	next  atomic.Pointer[segment[T]] // next segment in the linked list
	mu    sync.Mutex                 // protects write operations within this segment
	data  [segmentSize]T             // fixed-size array of events
	write atomic.Uint32              // current write position (0-63)
	read  uint32                     // current read position (consumer only)
}

// Queue is a multiple-producer, single-consumer queue for values of type T.
// Uses a linked list of fixed-size segments to minimize contention.
type Queue[T event.Event] struct {
	head atomic.Pointer[segment[T]] // producers write to head segment
	tail *segment[T]                // consumer reads from tail segment
	pool sync.Pool                  // recycles segments to reduce GC
}

func NewQueue[T event.Event]() *Queue[T] {
	q := &Queue[T]{}
	q.pool.New = func() any {
		return new(segment[T])
	}

	// Initialize with one empty segment
	seg := q.newSegment()
	q.head.Store(seg)
	q.tail = seg
	return q
}

// Push is safe for concurrent producers.
func (q *Queue[T]) Push(v T) {
	for {
		head := q.head.Load()

		// Try to write to current head segment under lock; revalidate head
		head.mu.Lock()
		if q.head.Load() != head {
			head.mu.Unlock()
			continue
		}

		writePos := head.write.Load()
		if writePos < segmentSize {
			// Space available in current segment
			head.data[writePos] = v
			head.write.Store(writePos + 1)
			head.mu.Unlock()
			return
		}

		// Segment is full; create and link a new head segment while holding lock
		newSeg := q.newSegment()
		newSeg.data[0] = v
		newSeg.write.Store(1)
		head.next.Store(newSeg)
		q.head.Store(newSeg)
		head.mu.Unlock()
		return
	}
}

// Pop must be called by a single goroutine.
// Returns zero value and false if empty.
func (q *Queue[T]) Pop() (T, bool) {
	var zero T

	for {
		// Check if current tail segment has data
		if q.tail.read < q.tail.write.Load() {
			val := q.tail.data[q.tail.read]
			q.tail.data[q.tail.read] = zero // clear for GC
			q.tail.read++
			return val, true
		}

		// Current segment is exhausted, try to move to next
		nextSeg := q.tail.next.Load()
		if nextSeg == nil {
			// No more segments, queue is empty
			return zero, false
		}

		// Move to next segment and recycle the current one
		oldTail := q.tail
		q.tail = nextSeg
		q.recycleSegment(oldTail)
	}
}

// Drain calls f for each available element, stopping early if f returns false.
// Single-consumer only.
func (q *Queue[T]) Drain(f func(T) bool) {
	for {
		v, ok := q.Pop()
		if !ok {
			return
		}
		if !f(v) {
			return
		}
	}
}

// Empty is an approximate check.
func (q *Queue[T]) Empty() bool {
	return q.tail.read >= q.tail.write.Load() && q.tail.next.Load() == nil
}

// LenApprox is a best-effort size estimate.
func (q *Queue[T]) LenApprox() int {
	count := 0

	// Count items in tail segment
	tailWrite := q.tail.write.Load()
	if tailWrite > q.tail.read {
		count += int(tailWrite - q.tail.read)
	}

	// Count items in linked segments
	for seg := q.tail.next.Load(); seg != nil; seg = seg.next.Load() {
		count += int(seg.write.Load())
	}

	return count
}

// Helpers

func (q *Queue[T]) newSegment() *segment[T] {
	seg := q.pool.Get().(*segment[T])
	seg.write.Store(0)
	seg.read = 0
	seg.next.Store(nil)
	return seg
}

func (q *Queue[T]) recycleSegment(seg *segment[T]) {
	// Clear the segment data for GC
	var zero T
	for i := range seg.data {
		seg.data[i] = zero
	}
	seg.write.Store(0)
	seg.read = 0
	seg.next.Store(nil)
	q.pool.Put(seg)
}
