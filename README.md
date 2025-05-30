<p align="center">
<img width="330" height="110" src=".github/logo.png" border="0" alt="kelindar/timeline">
<br>
<img src="https://img.shields.io/github/go-mod/go-version/kelindar/timeline" alt="Go Version">
<a href="https://pkg.go.dev/github.com/kelindar/timeline"><img src="https://pkg.go.dev/badge/github.com/kelindar/timeline" alt="PkgGoDev"></a>
<a href="https://goreportcard.com/report/github.com/kelindar/timeline"><img src="https://goreportcard.com/badge/github.com/kelindar/timeline" alt="Go Report Card"></a>
<a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"></a>
<a href="https://coveralls.io/github/kelindar/timeline"><img src="https://coveralls.io/repos/github/kelindar/timeline/badge.svg" alt="Coverage"></a>
</p>

## Timeline: High-Performance Task Scheduling in Go
This library provides a **high-performance, in-memory task scheduler** for Go, designed for precise and efficient time-based task management. It uses a 10ms resolution and a bucketing system for scalable scheduling, making it ideal for real-time and concurrent applications.

- **High Performance:** Optimized for rapid scheduling and execution of thousands of tasks per second, with minimal overhead.
- **Fine-Grained Precision:** Schedules tasks with 10ms accuracy, suitable for high-frequency or real-time workloads.
- **Efficient Memory Use:** Predictable, linear memory consumption thanks to its bucketing design.
- **Thread-Safe:** Safe for concurrent use, supporting multi-threaded scheduling and execution.

![demo](./.github/demo.gif)

**Use When:**
- ✅ Scheduling frequent, short-lived tasks (e.g., game loops, real-time updates).
- ✅ Requiring precise, low-latency task execution within a single Go process.
- ✅ Building systems where predictable memory and performance are critical.
- ✅ Needing a simple, dependency-free scheduler for in-process workloads.

**Not For:**
- ❌ Long-term scheduling (tasks days/weeks in advance).
- ❌ Applications with highly sporadic or infrequent task scheduling (due to required ticking).

## Quick Start

```go
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
```

It outputs:

```
Task executed at 04.400, elapsed=0s
Task executed at 05.000, elapsed=600ms
Task executed at 06.000, elapsed=1s
Task executed at 07.000, elapsed=1s
Task executed at 08.000, elapsed=1s
Task executed at 09.000, elapsed=1s
Task executed at 09.400, elapsed=5s
Task executed at 10.000, elapsed=1s
Task executed at 11.000, elapsed=1s
Task executed at 12.000, elapsed=1s
Task executed at 13.000, elapsed=1s
Task executed at 14.000, elapsed=1s
```

## Event Scheduling (Integration)

The [github.com/kelindar/timeline/emit](https://github.com/kelindar/timeline/tree/main/emit) sub-package seamlessly integrates the timeline scheduler with event-driven programming. It allows you to emit and subscribe to events with precise timing, making it ideal for applications that require both event-driven architectures and time-based scheduling.

```go
// Custom event type
type Message struct {
	Text string
}

// Type returns the type of the event for the dispatcher
func (Message) Type() uint32 {
	return 0x1
}

func main() {

	// Emit the event immediately
	event.Next(Message{Text: "Hello, World!"})

	// Emit the event every second
	event.Every(Message{Text: "Are we there yet?"}, 500*time.Millisecond)

	// Subscribe and Handle the Event
	cancel := event.On(func(ev Message, now time.Time, elapsed time.Duration) error {
		fmt.Printf("Received '%s' at %02d.%03d, elapsed=%v\n",
			ev.Text,
			now.Second(), now.UnixMilli()%1000, elapsed)
		return nil
	})
	defer cancel() // Remember to unsubscribe when done

	// Let the program run for a while to receive events
	time.Sleep(5 * time.Second)
}
```

The example above demonstrates how to create a custom event type, emit events, and subscribe to them using the timeline scheduler. It outputs:

```
Received 'Hello, World!' at 19.580, elapsed=0s
Received 'Are we there yet?' at 20.000, elapsed=420ms
Received 'Are we there yet?' at 20.500, elapsed=500ms
Received 'Are we there yet?' at 21.000, elapsed=500ms
Received 'Are we there yet?' at 21.500, elapsed=500ms
Received 'Are we there yet?' at 22.000, elapsed=500ms
Received 'Are we there yet?' at 22.500, elapsed=500ms
Received 'Are we there yet?' at 23.000, elapsed=500ms
Received 'Are we there yet?' at 23.500, elapsed=500ms
Received 'Are we there yet?' at 24.000, elapsed=500ms
Received 'Are we there yet?' at 24.500, elapsed=500ms
```

