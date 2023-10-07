package timeline

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	numBuckets = 100
)

// Task represents a task that can be scheduled.
type Task = func() bool

// job represents a scheduled task.
type job struct {
	Task
	Start Tick
	Every Tick
}

// bucket represents a bucket for a particular window of the second.
// TODO: this could be optimized with double buffering to reduce locking.
type bucket struct {
	mu    sync.Mutex
	queue []job
}

// Scheduler represents a task scheduler.
type Scheduler struct {
	next    atomic.Int64 // next tick
	buckets []*bucket
}

// New creates a new scheduler.
func New() *Scheduler {
	s := &Scheduler{
		buckets: make([]*bucket, numBuckets),
	}

	for i := 0; i < numBuckets; i++ {
		s.buckets[i] = &bucket{
			queue: make([]job, 0, 64),
		}
	}

	return s
}

// RunNext schedules a task to be processed during the next tick.
func (s *Scheduler) RunNext(task Task) {
	s.schedule(task, Tick(s.next.Load()), 0)
}

// RunAfter schedules a task to be processed after a given delay.
func (s *Scheduler) RunAfter(task Task, delay time.Duration) {
	s.schedule(task, Tick(s.next.Load())+durationOf(delay), 0)
}

// RunAt schedules a task to be processed at a given time.
func (s *Scheduler) RunAt(task Task, when time.Time) {
	s.schedule(task, TickOf(when), 0)
}

// RunEvery schedules a task to be processed at a given interval, starting
// immediately at the next tick.
func (s *Scheduler) RunEvery(task Task, interval time.Duration) {
	s.schedule(task, Tick(s.next.Load()), durationOf(interval))
}

// RunEveryAt schedules a task to be processed at a given interval,
// starting at a given time.
func (s *Scheduler) RunEveryAt(task Task, interval time.Duration, startTime time.Time) {
	s.schedule(task, TickOf(startTime), durationOf(interval))
}

// RunEveryAfter schedules a task to be processed at a given interval,
// starting after a given delay.
func (s *Scheduler) RunEveryAfter(task Task, interval, delay time.Duration) {
	s.schedule(task, Tick(s.next.Load())+durationOf(delay), durationOf(interval))
}

// ScheduleFunc schedules an event to be processed at a given time.
func (s *Scheduler) schedule(event Task, when, repeat Tick) {
	evt := job{
		Task:  event,
		Start: when,
		Every: repeat,
	}

	bucket := s.bucketOf(evt.Start)

	bucket.mu.Lock()
	bucket.queue = append(bucket.queue, evt)
	bucket.mu.Unlock()
}

// Seek advances the scheduler to a given time.
func (s *Scheduler) Seek(t time.Time) {
	s.next.Store(int64(TickOf(t)))
}

// Tick advances the scheduler to the next tick, processing all events
// and returning the current clock time.
func (s *Scheduler) Tick() time.Time {
	now := Tick(s.next.Add(1) - 1)
	bucket := s.bucketOf(now)
	offset := 0

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	for i, evt := range bucket.queue {
		if evt.Start > now { // scheduled for later
			bucket.queue[offset] = bucket.queue[i]
			offset++
			continue
		}

		// Process the task. If the task is recurrent, reschedule it
		if evt.Task(); evt.Every != 0 {
			s.schedule(evt.Task, now+evt.Every, evt.Every)
		}
	}

	// Truncate the current bucket to remove processed events
	bucket.queue = bucket.queue[:offset]
	return now.Time()
}

// bucketOf returns the bucket index for a given tick.
func (s *Scheduler) bucketOf(when Tick) *bucket {
	idx := int(when) % numBuckets
	return s.buckets[idx]
}

// durationOf computes a duration in terms of ticks.
func durationOf(t time.Duration) Tick {
	return Tick(t / resolution)
}
