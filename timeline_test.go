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
BenchmarkRun/next/1-24         	32000341	        37.56 ns/op	        32.00 million/op	       0 B/op	       0 allocs/op
BenchmarkRun/next/10-24        	 6282718	       191.8 ns/op	        62.83 million/op	       0 B/op	       0 allocs/op
BenchmarkRun/next/100-24       	  685710	      1746 ns/op	        68.57 million/op	       0 B/op	       0 allocs/op
BenchmarkRun/next/1000-24      	   70587	     17213 ns/op	        70.59 million/op	     105 B/op	       0 allocs/op
BenchmarkRun/next/10000-24     	    6966	    170543 ns/op	        69.66 million/op	   15890 B/op	       0 allocs/op
BenchmarkRun/next/100000-24    	     514	   2074903 ns/op	        51.40 million/op	 2738303 B/op	       4 allocs/op
BenchmarkRun/after/1-24        	31168992	        38.53 ns/op	        31.17 million/op	       0 B/op	       0 allocs/op
BenchmarkRun/after/10-24       	 6045394	       198.9 ns/op	        60.45 million/op	       0 B/op	       0 allocs/op
BenchmarkRun/after/100-24      	  685732	      1761 ns/op	        68.57 million/op	       0 B/op	       0 allocs/op
BenchmarkRun/after/1000-24     	   49078	     23361 ns/op	        48.58 million/op	    1220 B/op	       0 allocs/op
BenchmarkRun/after/10000-24    	    3808	    730699 ns/op	         7.252 million/op	 1168807 B/op	       0 allocs/op
BenchmarkRun/after/100000-24   	     369	   3436339 ns/op	         0.06827 million/op	12061832 B/op	       7 allocs/op
*/
func BenchmarkRun(b *testing.B) {
	work := func(time.Time, time.Duration) bool {
		counter.Add(1)
		return true
	}

	for _, size := range []int{1, 10, 100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("next/%d", size), func(b *testing.B) {
			counter.Store(0)
			s := New()
			b.ReportAllocs()
			b.ResetTimer()

			for n := 0; n < b.N; n++ {
				for i := 0; i < size; i++ {
					s.Run(work)
				}
				s.Tick()
			}

			b.ReportMetric(float64(counter.Load())/1000000, "million/op")
		})
	}

	for _, size := range []int{1, 10, 100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("after/%d", size), func(b *testing.B) {
			counter.Store(0)
			s := New()
			b.ReportAllocs()
			b.ResetTimer()

			for n := 0; n < b.N; n++ {
				for i := 0; i < size; i++ {
					s.RunAfter(work, time.Duration(10*i)*time.Millisecond)
				}
				s.Tick()
			}

			b.ReportMetric(float64(counter.Load())/1000000, "million/op")
		})
	}
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
