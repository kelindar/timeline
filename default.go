package timeline

import (
	"time"
)

// Default initializes a default timeline
var Default = New()

// Schedule schedules a task to be processed at a given time.
func Schedule(task Task, when time.Time) {
	Default.RunAt(task, when)
}
