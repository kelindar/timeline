package timeline

import (
	"time"
)

// Default initializes a default timeline
var Default = New()

// RunNext schedules a task to be processed during the next tick.
func RunNext(task Task) {
	Default.RunNext(task)
}

// RunAfter schedules a task to be processed after a given delay.
func RunAfter(task Task, delay time.Duration) {
	Default.RunAfter(task, delay)
}

// RunAt schedules a task to be processed at a given time.
func RunAt(task Task, when time.Time) {
	Default.RunAt(task, when)
}

// RunEvery schedules a task to be processed at a given interval, starting
// immediately at the next tick.
func RunEvery(task Task, interval time.Duration) {
	Default.RunEvery(task, interval)
}

// RunEveryAfter schedules a task to be processed at a given interval,
// starting at a given time.
func RunEveryAfter(task Task, interval time.Duration, startTime time.Time) {
	Default.RunEveryAfter(task, interval, startTime)
}
