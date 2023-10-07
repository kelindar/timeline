package timeline

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newScheduler(now time.Time) *Scheduler {
	s := New()
	s.Seek(now)
	return s
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

	assert.Equal(t, Counter(12), count)
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

	assert.Equal(t, Counter(12), count)
}

func TestRunEvery(t *testing.T) {
	now := time.Unix(0, 0)
	var count Counter

	s := newScheduler(now)
	s.RunEvery(count.Inc(), 10*time.Millisecond)
	s.RunEvery(count.Inc(), 30*time.Millisecond)

	for i := 0; i < 10; i++ {
		s.Tick()
	}

	assert.Equal(t, Counter(14), count)
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

	assert.Equal(t, Counter(2), count)
}

// ----------------------------------------- Log -----------------------------------------

// Log is a simple task that appends a string to a slice.
type Log []string

// Log returns a task that appends a string to the log.
func (l *Log) Log(s string) Task {
	return func() bool {
		*l = append(*l, s)
		return true
	}
}

// ----------------------------------------- Counter -----------------------------------------

type Counter int64

// Inc returns a task that increments the counter.
func (c *Counter) Inc() Task {
	return func() bool {
		*c++
		return true
	}
}
