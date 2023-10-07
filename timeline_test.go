package timeline

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

var counter atomic.Uint64

/*
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkRun/next/10-24         	 6315861	       190.6 ns/op	        63.16 million/op	       0 B/op	       0 allocs/op
BenchmarkRun/next/1000-24       	   70174	     17001 ns/op	        70.17 million/op	     106 B/op	       0 allocs/op
BenchmarkRun/next/100000-24     	     510	   2063725 ns/op	        51.00 million/op	 2759780 B/op	       4 allocs/op
BenchmarkRun/after/10-24        	 6045342	       197.3 ns/op	        60.45 million/op	       0 B/op	       0 allocs/op
BenchmarkRun/after/100-24       	  685722	      1759 ns/op	        68.57 million/op	       0 B/op	       0 allocs/op
BenchmarkRun/after/10000-24     	    3705	    693253 ns/op	         6.865 million/op	 1127870 B/op	       0 allocs/op
*/
func BenchmarkRun(b *testing.B) {
	work := func(time.Time, time.Duration) bool {
		counter.Add(1)
		return true
	}

	for _, size := range []int{10, 1000, 100000} {
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

	for _, size := range []int{10, 100, 10000} {
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
