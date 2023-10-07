package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kelindar/timeline"
)

func main() {
	// Initialize the scheduler and start the internal clock
	scheduler := timeline.New()
	cancel := scheduler.Start(context.Background())
	defer cancel() // Call this to stop the scheduler's internal clock

	// Define a task
	task := func(now time.Time, elapsed time.Duration) bool {
		fmt.Printf("Task executed at %d:%02d.%03d, elapsed=%v\n",
			now.Hour(), now.Second(), now.UnixMilli()%1000, elapsed)
		return true // return true to keep the task scheduled
	}

	// Schedule the task to run immediately
	scheduler.Run(task)

	// Schedule the task to run every second
	scheduler.RunEvery(task, 1*time.Second)

	// Schedule the task to run after 5 seconds
	scheduler.RunAfter(task, 5*time.Second)

	// Let the scheduler run for 10 seconds
	time.Sleep(10 * time.Second)
}
