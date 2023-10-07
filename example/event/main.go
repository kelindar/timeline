package main

import (
	"fmt"
	"time"

	"github.com/kelindar/timeline/event"
)

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
