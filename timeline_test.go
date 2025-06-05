// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package timeline

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

var counter atomic.Uint64

/*
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkRun/next-24         	11874015	       100.4 ns/op	        11.79 million/op	     117 B/op	       0 allocs/op
BenchmarkRun/after-24        	   10000	    471016 ns/op	         0.9101 million/op	    1469 B/op	       0 allocs/op
*/
func BenchmarkRun(b *testing.B) {
	work := func(time.Time, time.Duration) bool {
		counter.Add(1)
		return true
	}

	b.Run("next", func(b *testing.B) {
		counter.Store(0)
		s := New()
		s.Start(context.Background())
		b.ReportAllocs()
		b.ResetTimer()

		for n := 0; n < b.N; n++ {
			s.Run(work)
		}

		b.ReportMetric(float64(counter.Load())/1000000, "million/op")
	})

	b.Run("after", func(b *testing.B) {
		counter.Store(0)
		s := New()
		s.Start(context.Background())
		b.ReportAllocs()
		b.ResetTimer()

		for n := 0; n < b.N; n++ {
			for i := 0; i < 100; i++ {
				s.RunAfter(work, time.Duration(10*i)*time.Millisecond)
			}
		}

		b.ReportMetric(float64(counter.Load())/1000000, "million/op")
	})
}

func TestRunAt(t *testing.T) {
	now := time.Unix(0, 0)
	log := make(Log, 0, 8)

	s := newScheduler(now)
	s.RunAt(log.Log("Next 1"), now)
	s.RunAt(log.Log("Next 2"), now.Add(5*time.Millisecond))
	s.RunAt(log.Log("Future 1"), now.Add(495*time.Millisecond))
	s.RunAt(log.Log("Future 2"), now.Add(1600*time.Millisecond))

	for i := 0; i < 200; i++ {
		s.Tick()
	}

	assert.Equal(t, Log{
		"Next 1",
		"Next 2",
		"Future 1",
		"Future 2",
	}, log)
}

func TestRunAfter(t *testing.T) {
	now := time.Unix(0, 0)
	log := make(Log, 0, 8)

	s := newScheduler(now)
	s.RunAfter(log.Log("Next 1"), 0)
	s.RunAfter(log.Log("Next 2"), 5*time.Millisecond)
	s.RunAfter(log.Log("Future 1"), 495*time.Millisecond)
	s.RunAfter(log.Log("Future 2"), 1600*time.Millisecond)

	for i := 0; i < 200; i++ {
		s.Tick()
	}

	assert.Equal(t, Log{
		"Next 1",
		"Next 2",
		"Future 1",
		"Future 2",
	}, log)
}

func TestRunEveryAt(t *testing.T) {
	now := time.Unix(0, 0)
	var count Counter

	s := newScheduler(now)
	s.RunEveryAt(count.Inc(), 10*time.Millisecond, now)
	s.RunEveryAt(count.Inc(), 30*time.Millisecond, now.Add(50*time.Millisecond))

	for i := 0; i < 10; i++ {
		s.Tick()
	}

	assert.Equal(t, 12, count.Value())
}

func TestRunEveryAfter(t *testing.T) {
	now := time.Unix(0, 0)
	var count Counter

	s := newScheduler(now)
	s.RunEveryAfter(count.Inc(), 10*time.Millisecond, 0)
	s.RunEveryAfter(count.Inc(), 30*time.Millisecond, 50*time.Millisecond)

	for i := 0; i < 10; i++ {
		s.Tick()
	}

	assert.Equal(t, 12, count.Value())
}

func TestRunEvery10ms(t *testing.T) {
	now := time.Unix(0, 0)
	var count Counter

	s := newScheduler(now)
	s.RunEvery(count.Inc(), 10*time.Millisecond)

	for i := 0; i < 10; i++ {
		s.Tick()
	}

	assert.Equal(t, 9, count.Value())
}

func TestRunEvery1s(t *testing.T) {
	now := time.Unix(0, 0)
	var count Counter

	s := newScheduler(now)
	s.RunEvery(count.Inc(), 1*time.Second)

	for i := 0; i < 510; i++ {
		s.Tick()
	}

	assert.Equal(t, 5, count.Value())
}

func TestRun(t *testing.T) {
	now := time.Unix(0, 0)
	var count Counter

	s := newScheduler(now)
	s.Run(count.Inc())
	s.Run(count.Inc())

	for i := 0; i < 10; i++ {
		s.Tick()
	}

	assert.Equal(t, 2, count.Value())
}

func TestElapsed(t *testing.T) {
	s := New()

	var wg sync.WaitGroup
	wg.Add(3)
	s.RunEvery(func(now time.Time, elapsed time.Duration) bool {
		fmt.Printf("Tick at %02d.%03d, elapsed=%v\n",
			now.Second(), now.UnixMilli()%1000, elapsed)
		assert.Equal(t, 10*time.Millisecond, elapsed)
		wg.Done()
		return true
	}, 10*time.Millisecond)

	s.Tick()
	s.Tick()
	s.Tick()
	s.Tick()
	wg.Wait()
}

func TestTickOf(t *testing.T) {
	tc := map[tick]time.Duration{
		0:      0,
		1:      10 * time.Millisecond,
		2:      20 * time.Millisecond,
		10:     100 * time.Millisecond,
		100:    time.Second,
		101:    time.Second + 10*time.Millisecond,
		360000: time.Hour,
	}

	for expect, duration := range tc {
		assert.Equal(t, expect, tickOf(time.Unix(0, int64(duration))))
	}
}

func TestStart(t *testing.T) {
	s := New()
	defer s.Start(context.Background())()

	var count Counter
	s.RunAfter(count.Inc(), 30*time.Millisecond)
	s.Run(count.Inc())
	s.Run(count.Inc())

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 3, count.Value())
}

func TestJobSize(t *testing.T) {
	size := unsafe.Sizeof(job{})
	assert.Equal(t, 24, int(size))
}

func TestRunDuringTickDeadlocks(t *testing.T) {
	s := newScheduler(time.Unix(0, 0))

	s.Run(func(time.Time, time.Duration) bool {
		s.Run(func(time.Time, time.Duration) bool { return false })
		return false
	})

	done := make(chan struct{})
	go func() {
		s.Tick()
		close(done)
	}()

	select {
	case <-done:
		// Success! Tick() completed without deadlock
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Tick() deadlocked - did not complete within timeout")
	}
}

func TestNestedSchedulingScenarios(t *testing.T) {
	s := newScheduler(time.Unix(0, 0))
	var execCount Counter

	// Test 1: Task scheduling another task in the same bucket
	s.Run(func(time.Time, time.Duration) bool {
		execCount.Inc()(time.Now(), 0)
		s.Run(func(time.Time, time.Duration) bool {
			execCount.Inc()(time.Now(), 0)
			return false
		})
		return false
	})

	// Test 2: Task scheduling a recurring task
	s.Run(func(time.Time, time.Duration) bool {
		execCount.Inc()(time.Now(), 0)
		s.RunEvery(func(time.Time, time.Duration) bool {
			execCount.Inc()(time.Now(), 0)
			return false // run only once
		}, 10*time.Millisecond)
		return false
	})

	// Execute multiple ticks to process all tasks
	for i := 0; i < 5; i++ {
		s.Tick()
	}

	// Should have executed: 2 initial tasks + 2 nested tasks = 4 total
	assert.Equal(t, 4, execCount.Value())
}

// ----------------------------------------- Log -----------------------------------------

// Log is a simple task that appends a string to a slice.
type Log []string

// Log returns a task that appends a string to the log.
func (l *Log) Log(s string) Task {
	return func(time.Time, time.Duration) bool {
		*l = append(*l, s)
		return true
	}
}

// ----------------------------------------- Counter -----------------------------------------

type Counter int64

// Value returns the current value of the counter.
func (c *Counter) Value() int {
	return int(atomic.LoadInt64((*int64)(c)))
}

// Inc returns a task that increments the counter.
func (c *Counter) Inc() Task {
	return func(time.Time, time.Duration) bool {
		atomic.AddInt64((*int64)(c), 1)
		return true
	}
}

// ----------------------------------------- Scheduler -----------------------------------------

func newScheduler(now time.Time) *Scheduler {
	s := New()
	s.Seek(now)
	return s
}
