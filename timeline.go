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
	Time   Tick
	Repeat Tick
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

func (tl *Timeline) Delay(event Event, delay time.Duration) {
	// TODO: avoid using time.Now()
	tl.schedule(event, TickOf(time.Now().Add(delay)), 0)
}

// ScheduleFunc schedules an event to be processed at a given time.
func (tl *Timeline) Schedule(event Event, when time.Time) {
	tl.schedule(event, TickOf(when), 0)
}

// Repeat schedules an event to be processed at a given interval, starting
// immediately at the next tick.
func (tl *Timeline) Repeat(event Event, interval time.Duration) {
	tl.schedule(event, Tick(tl.next.Load()), durationOf(interval))
}

// Repeat schedules an event to be processed at a given interval, starting
// at a given time.
func (tl *Timeline) RepeatAfter(event Event, interval time.Duration, startTime time.Time) {
	tl.schedule(event, TickOf(startTime), durationOf(interval))
}

// Next schedules an event to be processed during the next tick.
func (tl *Timeline) Next(event Event) {
	tl.schedule(event, Tick(tl.next.Load()), 0)
}

// ScheduleFunc schedules an event to be processed at a given time.
func (tl *Timeline) schedule(event Event, when, repeat Tick) {
	evt := plan{
		Event:  event,
		Time:   when,
		Repeat: repeat,
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

		// If the event has a non-zero interval, update its execution time
		if evt.Repeat != 0 {
			evt.Time = now + evt.Repeat
			bucket.queue[offset] = evt
			offset++
		}
	}

	// Truncate the current bucket to remove processed events
	bucket.queue = bucket.queue[:offset]
}

// bucketOf returns the bucket index for a given tick.
func (tl *Timeline) bucketOf(when Tick) *bucket {
	idx := int(when) % numBuckets
	return tl.buckets[idx]
}

// durationOf computes a duration in terms of ticks.
func durationOf(t time.Duration) Tick {
	return Tick(t / resolution)
}
