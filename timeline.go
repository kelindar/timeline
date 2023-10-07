package timeline

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	numBuckets = 100
)

// Event represents an event that can be scheduled.
type Event interface {
	Execute()
}

type plan struct {
	Event
	Time Tick
}

// Timeline represents a timeline of events.
type Timeline struct {
	next    atomic.Int64 // next tick
	buckets []*bucket
}

// bucket represents a bucket for a particular window of the second.
// TODO: this could be optimized with double buffering to reduce locking.
type bucket struct {
	mu    sync.Mutex
	queue []plan
}

// New creates a new timeline.
func New() *Timeline {
	tl := &Timeline{
		buckets: make([]*bucket, numBuckets),
	}

	for i := 0; i < numBuckets; i++ {
		tl.buckets[i] = &bucket{
			queue: make([]plan, 0, 64),
		}
	}

	return tl
}

// ScheduleFunc schedules an event to be processed at a given time.
func (tl *Timeline) Schedule(event Event, when time.Time) {
	tl.schedule(event, TickOf(when))
}

// Next schedules an event to be processed during the next tick.
func (tl *Timeline) Next(event Event) {
	tl.schedule(event, Tick(tl.next.Load()))
}

// ScheduleFunc schedules an event to be processed at a given time.
func (tl *Timeline) schedule(event Event, when Tick) {
	evt := plan{
		Time:  when,
		Event: event,
	}

	bucket := tl.bucketOf(evt.Time)

	bucket.mu.Lock()
	bucket.queue = append(bucket.queue, evt)
	bucket.mu.Unlock()
}

// Tick processes all events that are due for processing at the given time.
func (tl *Timeline) Tick(now Tick) {
	tl.next.Store(int64(now) + 1)

	bucket := tl.bucketOf(now)
	offset := 0

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	for i, evt := range bucket.queue {
		if evt.Time > now { // scheduled for later
			bucket.queue[offset] = bucket.queue[i]
			offset++
			continue
		}

		// Process the event
		evt.Execute()
	}

	// Truncate the current bucket to remove processed events
	bucket.queue = bucket.queue[:offset]
}

// bucketOf returns the bucket index for a given tick.
func (tl *Timeline) bucketOf(when Tick) *bucket {
	idx := int(when) % numBuckets
	return tl.buckets[idx]
}
