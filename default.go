package timeline

import (
	"time"
)

// defaultTimeline initializes a default timeline
var defaultTimeline = New()

// Schedule schedules an event to be processed at a given time.
func Schedule[T Event](event T, when time.Time) {
	defaultTimeline.Schedule(event, when)
}
