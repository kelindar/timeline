## Overview

This library provides a simple and efficient way to schedule and manage tasks based on time. It offers a fine-grained resolution of 10 milliseconds and uses a bucketing system to efficiently manage scheduled tasks. The library is designed to be thread-safe and can handle concurrent scheduling and execution of tasks.

## Features

- **Task Scheduling**: Schedule tasks to run immediately, at a specific time, after a delay, or at regular intervals.
- **Fine-grained Resolution**: Tasks are scheduled with a resolution of 10 milliseconds.
- **Efficient Bucketing System**: Uses a bucketing system to efficiently manage and execute scheduled tasks.
- **Thread-safe**: Designed to handle concurrent scheduling and execution of tasks.

## Usage

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

```json
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
