package main

import (
	"fmt"
	"time"

	"github.com/kelindar/timeline/emit"
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
	emit.Next(Message{Text: "Hello, World!"})

	// Emit the event every second
	emit.Every(Message{Text: "Are we there yet?"}, 500*time.Millisecond)

	// Subscribe and Handle the Event
	cancel := emit.On[Message](func(ev Message, now time.Time, elapsed time.Duration) error {
		fmt.Printf("Received '%s' at %02d.%03d, elapsed=%v\n",
			ev.Text,
			now.Second(), now.UnixMilli()%1000, elapsed)
		return nil
	})
	defer cancel() // Remember to unsubscribe when done

	// Let the program run for a while to receive events
	time.Sleep(5 * time.Second)
}
