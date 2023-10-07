package timeline

import (
	"sync/atomic"
	"time"
)

const (
	resolution = 10 * time.Millisecond
)

// ----------------------------------------- Clock -----------------------------------------

// Clock represents a game clock
type Clock struct {
	last int64 // Last updated time, in unix nano
	now  func() int64
}

// newClock creates a new clock
func newClock() *Clock {
	return &Clock{
		last: time.Now().UTC().UnixNano(),
		now: func() int64 {
			return time.Now().UTC().UnixNano()
		},
	}
}

// Elapsed computes the time between now and the last update
func (c *Clock) Elapsed() time.Duration {
	return time.Duration(c.now() - atomic.LoadInt64(&c.last)).Round(1 * time.Millisecond)
}

// Tick updates the current clock and returns the elapsed time
func (c *Clock) Tick() time.Duration {
	now := c.now()
	dt := time.Duration(now - atomic.LoadInt64(&c.last)).Round(100 * time.Millisecond)
	atomic.StoreInt64(&c.last, now)
	return dt
}

// ----------------------------------------- Tick -----------------------------------------

// Tick represents a point in time, rounded up to the resolution of the clock.
type Tick int64

// TickOf returns the time rounded up to the resolution of the clock.
func TickOf(t time.Time) Tick {
	return Tick(t.UnixNano() / int64(resolution))
}

/*

func main() {
	cq := New()

	// Schedule some events
	now := time.Now()
	event1 := Event{Data: "Immediate Event 1"}
	event2 := Event{Data: "Immediate Event 2"}
	event3 := Event{Data: "Future Event 1"}
	event4 := Event{Data: "Future Event 2"}

	cq.Schedule(event1, now)
	cq.Schedule(event2, now.Add(5*time.Millisecond))
	cq.Schedule(event3, now.Add(495*time.Millisecond))
	cq.Schedule(event4, now.Add(1600*time.Millisecond))

	// Cancel event2
	cq.Cancel(&event2)

	fmt.Println("Processing events...")

	// Process events every 10ms (simulating the tick)
	ticker := time.NewTicker(initialWidth)
	count := 0
	for range ticker.C {
		fmt.Println("Tick", count)
		cq.ProcessCurrentBucket()

		// For demonstration purposes, stop after 2 seconds
		count++
		if count == 20 {
			ticker.Stop()
			break
		}
	}

	fmt.Println("Done")
}
*/
