package timeline

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

const (
	resolution = 10 * time.Millisecond
	numBuckets = int(1 * time.Second / resolution)
)

// Task represents a task that can be scheduled.
type Task = func(time.Time) bool

// job represents a scheduled task.
type job struct {
	Task
	Start tick
	Every tick
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

// Run schedules a task to be processed during the next tick.
func (s *Scheduler) Run(task Task) {
	s.schedule(task, tick(s.next.Load()), 0)
}

// RunAt schedules a task to be processed at a given time.
func (s *Scheduler) RunAt(task Task, when time.Time) {
	s.schedule(task, tickOf(when), 0)
}

// RunAfter schedules a task to be processed after a given delay.
func (s *Scheduler) RunAfter(task Task, delay time.Duration) {
	s.schedule(task, tick(s.next.Load())+durationOf(delay), 0)
}

// RunEvery schedules a task to be processed at a given interval, starting
// immediately at the next tick.
func (s *Scheduler) RunEvery(task Task, interval time.Duration) {
	s.schedule(task, tick(s.next.Load()), durationOf(interval))
}

// RunEveryAt schedules a task to be processed at a given interval,
// starting at a given time.
func (s *Scheduler) RunEveryAt(task Task, interval time.Duration, startTime time.Time) {
	s.schedule(task, tickOf(startTime), durationOf(interval))
}

// RunEveryAfter schedules a task to be processed at a given interval,
// starting after a given delay.
func (s *Scheduler) RunEveryAfter(task Task, interval, delay time.Duration) {
	s.schedule(task, tick(s.next.Load())+durationOf(delay), durationOf(interval))
}

// ScheduleFunc schedules an event to be processed at a given time.
func (s *Scheduler) schedule(event Task, when, repeat tick) {
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
	s.next.Store(int64(tickOf(t)))
}

// Tick advances the scheduler to the next tick, processing all events
// and returning the current clock time.
func (s *Scheduler) Tick() time.Time {
	tickNow := tick(s.next.Add(1) - 1)
	timeNow := tickNow.Time()
	bucket := s.bucketOf(tickNow)
	offset := 0

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	for i, evt := range bucket.queue {
		if evt.Start > tickNow { // scheduled for later
			bucket.queue[offset] = bucket.queue[i]
			offset++
			continue
		}

		// Process the task. If the task is recurrent, reschedule it
		if evt.Task(timeNow); evt.Every != 0 {
			s.schedule(evt.Task, tickNow+evt.Every, evt.Every)
		}
	}

	// Truncate the current bucket to remove processed events
	bucket.queue = bucket.queue[:offset]
	return tickNow.Time()
}

// bucketOf returns the bucket index for a given tick.
func (s *Scheduler) bucketOf(when tick) *bucket {
	idx := int(when) % numBuckets
	return s.buckets[idx]
}

// ----------------------------------------- Clock -----------------------------------------

// Start starts the scheduler internal clock.
func (s *Scheduler) Start(ctx context.Context) context.CancelFunc {
	interval := resolution
	ctx, cancel := context.WithCancel(ctx)

	// Calculate the time until the next 10ms boundary
	now := time.Now()
	next := now.Truncate(interval).Add(interval)
	wait := next.Sub(now)

	// Wait until the next resolution boundary
	time.Sleep(wait)

	// Start the ticker
	ticker := time.NewTicker(interval)
	s.Tick()
	go func() {
		for {
			select {
			case <-ticker.C:
				s.Tick()
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	return cancel
}

// ----------------------------------------- Tick -----------------------------------------

// tick represents a point in time, rounded up to the resolution of the clock.
type tick int64

// Time returns the time of the tick.
func (t tick) Time() time.Time {
	return time.Unix(0, int64(t)*int64(resolution))
}

// tickOf returns the time rounded up to the resolution of the clock.
func tickOf(t time.Time) tick {
	return tick(t.UnixNano() / int64(resolution))
}

// durationOf computes a duration in terms of ticks.
func durationOf(t time.Duration) tick {
	return tick(t / resolution)
}
