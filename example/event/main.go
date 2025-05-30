package main

import (
	"fmt"
	"time"

	"github.com/kelindar/timeline/emit"
)

// Say event type for dialogue
type Say struct {
	Character string
	Line      string
}

// Type returns the type of the Say event
func (Say) Type() uint32 { return 0x1 }

// Do event type for actions (with closures)
type Do struct {
	Action func()
}

// Type returns the type of the Do event
func (Do) Type() uint32 { return 0x2 }

func main() {
	// Subscribe to Say events for dialogue
	sayCancel := emit.On(func(ev Say, now time.Time, elapsed time.Duration) error {
		fmt.Printf("[%02d.%03d] %s: %s\n",
			now.Second(), now.UnixMilli()%1000,
			ev.Character, ev.Line,
		)
		return nil
	})
	defer sayCancel()

	// Subscribe to Do events for actions
	doCancel := emit.On(func(ev Do, now time.Time, elapsed time.Duration) error {
		ev.Action() // Execute the closure
		return nil
	})
	defer doCancel()

	// Schedule the dialogue using emit timing functions
	scheduleDialogue()

	// Let the dialogue play out over about 10 seconds
	time.Sleep(10 * time.Second)
}

func scheduleDialogue() {
	// Demonstrate emit.Next() - immediate events
	emit.Next(Say{Character: "🎵 Narrator", Line: "Journey begins..."})

	// Demonstrate emit.After() - scheduled delays
	emit.After(Say{Character: "🐴 Donkey", Line: "Are we there yet?"}, 500*time.Millisecond)
	emit.After(Say{Character: "👹 Shrek", Line: "No."}, 1*time.Second)

	// Demonstrate emit.EveryAfter() with cancellation - recurring annoyance!
	donkeyNagging := emit.EveryAfter(Say{Character: "🐴 Donkey", Line: "Are we there YET?"}, 900*time.Millisecond, 2*time.Second)

	// Responses to the recurring annoyance (offset to avoid collisions)
	emit.After(Say{Character: "👹 Shrek", Line: "NO!"}, 2500*time.Millisecond)
	emit.After(Say{Character: "👸 Fiona", Line: "NOT YET!"}, 3200*time.Millisecond)
	emit.After(Say{Character: "👹 Shrek", Line: "STOP ASKING!"}, 4100*time.Millisecond)

	// Stop Donkey's nagging when Shrek gets really mad - use Do event with closure
	emit.After(Do{Action: func() {
		donkeyNagging() // Cancel the recurring event
		fmt.Println("🎭 [Donkey stops nagging]")
	}}, 5*time.Second)

	// Dramatic pause with background sounds
	emit.After(Say{Character: "🎵 Narrator", Line: "[awkward silence]"}, 5500*time.Millisecond)
	lipPopping := emit.EveryAfter(Say{Character: "🎵 Narrator", Line: "*pop*"}, 600*time.Millisecond, 6*time.Second)

	// Final explosion
	emit.After(Say{Character: "👹 Shrek", Line: "THAT'S IT!!!"}, 8*time.Second)

	// Stop the lip popping when Fiona announces arrival - use Do event with closure
	emit.After(Do{Action: func() {
		lipPopping() // Cancel the lip popping
		fmt.Println("🎭 [Background sounds stop]")
	}}, 8500*time.Millisecond)

	emit.After(Say{Character: "👸 Fiona", Line: "We're here!"}, 8500*time.Millisecond)
	emit.After(Say{Character: "🐴 Donkey", Line: "Finally! 🎉"}, 9*time.Second)
}
