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

This library provides a simple and efficient way to schedule and manage tasks based on time. It offers a fine-grained resolution of 10 milliseconds and uses a bucketing system to efficiently manage scheduled tasks. The library is designed to be thread-safe and can handle concurrent scheduling and execution of tasks.

### Advantages

1. **High Performance**: This library is optimized for speed, handling a large number of tasks with minimal overhead. For instance, it's ideal for real-time game servers where tasks like player movements or AI decisions need frequent scheduling.

2. **Fine-grained Resolution**: With its 10ms resolution, this package offers precise scheduling. This resolution is useful for applications where tasks need to be scheduled at a high frequency.

3. **Efficient Memory Management**: The library's bucketing system ensures linear and predictable memory consumption. This efficiency is beneficial in cloud environments where memory usage impacts costs.

4. **Thread-safe**: Timeline is designed for concurrent scheduling and execution, making it suitable for multi-threaded applications like web servers handling simultaneous requests.

### Disadvantages

1. **Not Suitable for Long-term Scheduling**: This library is optimized for short-term tasks. It's not intended for tasks scheduled days or weeks in advance, making it less ideal for applications like calendar reminders.

2. **Requires Active Ticking**: The library needs active ticking (via the Tick method) to process tasks. This design might not be suitable for scenarios with sporadic task scheduling.

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
Task executed at 22:04.400, elapsed=0s
Task executed at 22:05.000, elapsed=600ms
Task executed at 22:06.000, elapsed=1s
Task executed at 22:07.000, elapsed=1s
Task executed at 22:08.000, elapsed=1s
Task executed at 22:09.000, elapsed=1s
Task executed at 22:09.400, elapsed=5s
Task executed at 22:10.000, elapsed=1s
Task executed at 22:11.000, elapsed=1s
Task executed at 22:12.000, elapsed=1s
Task executed at 22:13.000, elapsed=1s
Task executed at 22:14.000, elapsed=1s
```

## Benchmarks

| Type  | Input Size | Nanoseconds/Op | Million Run/Sec | Allocs/Op |
| ----- | ---------- | -------------- | --------------- | --------- |
| next  | 1          | 37.56          | 32.0 Million    | 0         |
| next  | 10         | 191.8          | 62.83 Million   | 0         |
| next  | 100        | 1746.0         | 68.57 Million   | 0         |
| next  | 1000       | 17213.0        | 70.59 Million   | 0         |
| next  | 10000      | 170543.0       | 69.66 Million   | 0         |
| next  | 100000     | 2074903.0      | 51.4 Million    | 4         |
| after | 1          | 38.53          | 31.17 Million   | 0         |
| after | 10         | 198.9          | 60.45 Million   | 0         |
| after | 100        | 1761.0         | 68.57 Million   | 0         |
| after | 1000       | 23361.0        | 48.58 Million   | 0         |
| after | 10000      | 730699.0       | 7.252 Million   | 0         |
| after | 100000     | 3436339.0      | 0.06827 Million | 7         |

## Event Package for Timeline

The `event`` sub-package seamlessly integrates the timeline scheduler with event-driven programming. It allows you to emit and subscribe to events with precise timing, making it ideal for applications that require both event-driven architectures and time-based scheduling.

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
	event.Emit(Message{Text: "Hello, World!"})

	// Emit the event every second
	event.EmitEvery(Message{Text: "Are we there yet?"}, 1*time.Second)

	// Subscribe and Handle the Event
	cancel := event.On[Message](func(ev event.Event[Message]) {
		fmt.Printf("Received '%s' at %02d.%03d, elapsed=%v\n",
			ev.Data.Text,
			ev.Time.Second(), ev.Time.UnixMilli()%1000, ev.Elapsed)
	})
	defer cancel() // Remember to unsubscribe when done

	// Let the program run for a while to receive events
	time.Sleep(5 * time.Second)
}
```
