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

// Task defines a scheduled function. 'now' is the execution time, and 'elapsed'
// indicates the time since the last schedule or execution.  The return value of
// the function is a boolean. If the task returns 'true', it indicates that the
// task should continue to be scheduled for future execution based on its
// interval. Returning 'false' implies that the task should not be executed again.
type Task = func(now time.Time, elapsed time.Duration) bool

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

// Scheduler manages and executes scheduled tasks.
type Scheduler struct {
	next    atomic.Int64 // next tick
	buckets []*bucket
}

// New initializes and returns a new Scheduler.
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

// Run schedules a task for the next tick.
func (s *Scheduler) Run(task Task) {
	s.schedule(task, tick(s.next.Load()), 0)
}

// RunAt schedules a task for a specific 'when' time.
func (s *Scheduler) RunAt(task Task, when time.Time) {
	s.schedule(task, tickOf(when), 0)
}

// RunAfter schedules a task to run after a 'delay'.
func (s *Scheduler) RunAfter(task Task, delay time.Duration) {
	s.schedule(task, tick(s.next.Load())+durationOf(delay), 0)
}

// RunEvery schedules a task to run at 'interval' intervals, starting immediately.
func (s *Scheduler) RunEvery(task Task, interval time.Duration) {
	s.schedule(task, tick(s.next.Load()), durationOf(interval))
}

// RunEveryAt schedules a task to run at 'interval' intervals, starting at 'startTime'.
func (s *Scheduler) RunEveryAt(task Task, interval time.Duration, startTime time.Time) {
	s.schedule(task, tickOf(startTime), durationOf(interval))
}

// RunEveryAfter schedules a task to run at 'interval' intervals after a 'delay'.
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

// Tick processes tasks for the current time and advances the internal clock.
func (s *Scheduler) Tick() time.Time {
	tickNow := tick(s.next.Add(1) - 1)
	timeNow := tickNow.Time()
	bucket := s.bucketOf(tickNow)
	offset := 0

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	for i, task := range bucket.queue {
		if task.Start > tickNow { // scheduled for later
			bucket.queue[offset] = bucket.queue[i]
			offset++
			continue
		}

		// Process the task. If the task is recurrent, reschedule it
		elapsed := tickNow - task.Start
		if task.Task(timeNow, elapsed.Duration()); task.Every != 0 {
			s.schedule(task.Task, tickNow+task.Every, task.Every)
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

// Start begins the scheduler's internal clock, aligning with the specified
// 'interval'. It returns a cancel function to stop the clock.
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

// Time converts the tick to a timestamp.
func (t tick) Time() time.Time {
	return time.Unix(0, int64(t)*int64(resolution))
}

// Duration converts the tick to a duration.
func (t tick) Duration() time.Duration {
	return time.Duration(t) * resolution
}

// tickOf returns the time rounded up to the resolution of the clock.
func tickOf(t time.Time) tick {
	return tick(t.UnixNano() / int64(resolution))
}

// durationOf computes a duration in terms of ticks.
func durationOf(t time.Duration) tick {
	return tick(t / resolution)
}
