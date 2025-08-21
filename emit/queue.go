package emit

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/kelindar/event"
	"github.com/kelindar/timeline"
)

const segmentSize = 64

var _ timeline.Task = (*queue[fault])(nil).Drain

// segment represents a fixed-size buffer that holds events
type segment[T any] struct {
	next  atomic.Pointer[segment[T]] // next segment in the linked list
	mu    sync.Mutex                 // protects write operations within this segment
	data  [segmentSize]T             // fixed-size array of events
	write atomic.Uint32              // current write position (0-63)
	read  uint32                     // current read position (consumer only)
}

// queue is a multiple-producer, single-consumer queue
type queue[T event.Event] struct {
	head atomic.Pointer[segment[T]] // producers write to head segment
	tail *segment[T]                // consumer reads from tail segment
	pool sync.Pool                  // recycles segments to reduce GC
}

// newQueue creates a new queue
func newQueue[T event.Event]() *queue[T] {
	q := &queue[T]{}
	q.pool.New = func() any {
		return new(segment[T])
	}

	// Initialize with one empty segment
	seg := q.borrow()
	q.head.Store(seg)
	q.tail = seg
	return q
}

// Push is safe for concurrent producers.
func (q *queue[T]) Push(v T) {
	for {
		head := q.head.Load()

		// Try to write to current head segment under lock; revalidate head
		head.mu.Lock()
		if q.head.Load() != head {
			head.mu.Unlock()
			continue
		}

		// Space available in current segment
		if writeAt := head.write.Load(); writeAt < segmentSize {
			head.data[writeAt] = v
			head.write.Store(writeAt + 1)
			head.mu.Unlock()
			return
		}

		// Segment is full; create and link a new head seg while holding lock
		seg := q.borrow()
		seg.data[0] = v
		seg.write.Store(1)
		head.next.Store(seg)
		q.head.Store(seg)
		head.mu.Unlock()
		return
	}
}

// Drain is called by the scheduler to publish events, single-consumer only.
func (q *queue[T]) Drain(now time.Time, elapsed time.Duration) bool {
	var zero T
	for {
		seg := q.tail
		idx := seg.write.Load()

		// Process all items in this segment, until the write index
		for seg.read < idx {
			val := seg.data[seg.read]
			seg.data[seg.read] = zero // clear for GC
			seg.read++

			// Publish the event
			event.Publish(event.Default, signal[T]{
				Data:    val,
				Time:    tickOf(now),
				Elapsed: durationOf(elapsed),
			})
		}

		// Current segment is exhausted
		next := seg.next.Load()
		if next == nil {
			return true // empty
		}

		// Move to next segment and reset the current one
		q.tail = next
		q.reset(seg)
	}
}

func (q *queue[T]) borrow() *segment[T] {
	seg := q.pool.Get().(*segment[T])
	seg.write.Store(0)
	seg.read = 0
	seg.next.Store(nil)
	return seg
}

func (q *queue[T]) reset(seg *segment[T]) {
	seg.write.Store(0)
	seg.read = 0
	seg.next.Store(nil)
	q.pool.Put(seg)
}
