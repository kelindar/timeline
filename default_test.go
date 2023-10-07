package timeline

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunAt(t *testing.T) {
	now := time.Unix(0, 0)
	log := make(Log, 0, 8)

	RunAt(log.Log("Next 1"), now)
	RunAt(log.Log("Next 2"), now.Add(5*time.Millisecond))
	RunAt(log.Log("Future 1"), now.Add(495*time.Millisecond))
	RunAt(log.Log("Future 2"), now.Add(1600*time.Millisecond))

	for ts := Tick(0); ts < 200; ts++ {
		Default.Tick(ts)
	}

	assert.Equal(t, Log{
		"Next 1",
		"Next 2",
		"Future 1",
		"Future 2",
	}, log)
}

func TestRunAfter(t *testing.T) {
	log := make(Log, 0, 8)

	RunAfter(log.Log("Next 1"), 0)
	RunAfter(log.Log("Next 2"), 5*time.Millisecond)
	RunAfter(log.Log("Future 1"), 495*time.Millisecond)
	RunAfter(log.Log("Future 2"), 1600*time.Millisecond)

	now := TickOf(time.Now())
	for ts := now; ts < now+200; ts++ {
		Default.Tick(ts)
	}

	assert.Equal(t, Log{
		"Next 1",
		"Next 2",
		"Future 1",
		"Future 2",
	}, log)
}

func TestRunEveryAfter(t *testing.T) {
	now := time.Unix(0, 0)
	var count Counter

	RunEveryAfter(count.Inc(), 10*time.Millisecond, now)
	RunEveryAfter(count.Inc(), 30*time.Millisecond, now.Add(50*time.Millisecond))

	for ts := Tick(0); ts < 10; ts++ {
		Default.Tick(ts)
	}

	assert.Equal(t, Counter(12), count)
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
