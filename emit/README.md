## Event Package for Timeline

The event package seamlessly integrates the timeline scheduler with event-driven programming. It allows you to emit and subscribe to events with precise timing, making it ideal for applications that require both event-driven architectures and time-based scheduling.

## Quick Start

Let's dive right in with a simple example to get you started with the event package.

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
	event.Every(Message{Text: "Are we there yet?"}, 1*time.Second)

	// Subscribe and Handle the Event
	cancel := event.On[Message](func(ev Message, now time.Time, elapsed time.Duration) error {
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

You will see similar output, with 'Are we there yet?' being emitted every second, and 'Hello, World!' being emitted immediately.

```
Received 'Hello, World!' at 21.060, elapsed=0s
Received 'Are we there yet?' at 22.000, elapsed=940ms
Received 'Are we there yet?' at 23.000, elapsed=1s
Received 'Are we there yet?' at 24.000, elapsed=1s
Received 'Are we there yet?' at 25.000, elapsed=1s
Received 'Are we there yet?' at 26.000, elapsed=1s
```
