package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/kelindar/timeline"
)

func main() {
	// Create and start the timeline scheduler
	scheduler := timeline.New()
	cancel := scheduler.Start(context.Background())
	defer cancel()

	// Schedule the dialogue using timeline directly
	scheduleDialogue(scheduler)

	// Let the dialogue play out over about 10 seconds
	time.Sleep(10 * time.Second)
}

func scheduleDialogue(scheduler *timeline.Scheduler) {
	say := func(character, line string) timeline.Task {
		return func(now time.Time, elapsed time.Duration) bool {
			fmt.Printf("[%02d.%03d] %s: %s\n", now.Second(), now.UnixMilli()%1000, character, line)
			return false // Don't repeat
		}
	}

	// Cancellable repeating say task
	sayEvery := func(character, line string) (timeline.Task, func()) {
		var stop atomic.Bool
		task := func(now time.Time, elapsed time.Duration) (repeat bool) {
			if repeat = !stop.Load(); repeat {
				fmt.Printf("[%02d.%03d] %s: %s\n", now.Second(), now.UnixMilli()%1000, character, line)
			}
			return
		}
		cancel := func() { stop.Store(true) }
		return task, cancel
	}

	// Immediate events - run next tick
	scheduler.Run(say("🎵 Narrator", "Journey begins..."))

	// Scheduled delays using RunAfter
	scheduler.RunAfter(say("🐴 Donkey", "Are we there yet?"), 500*time.Millisecond)
	scheduler.RunAfter(say("👹 Shrek", "No."), 1*time.Second)

	// Recurring events with RunEveryAfter - recurring annoyance!
	donkeyNagging, cancelDonkey := sayEvery("🐴 Donkey", "Are we there YET?")
	scheduler.RunEveryAfter(donkeyNagging, 900*time.Millisecond, 2*time.Second)

	// Responses to the recurring annoyance (offset to avoid collisions)
	scheduler.RunAfter(say("👹 Shrek", "NO!"), 2500*time.Millisecond)
	scheduler.RunAfter(say("👸 Fiona", "NOT YET!"), 3200*time.Millisecond)
	scheduler.RunAfter(say("👹 Shrek", "STOP ASKING!"), 4100*time.Millisecond)

	// Stop Donkey's nagging when Shrek gets really mad
	scheduler.RunAfter(func(now time.Time, elapsed time.Duration) bool {
		cancelDonkey()
		fmt.Println("🎭 [Donkey stops nagging]")
		return false
	}, 5*time.Second)

	// Dramatic pause with background sounds
	scheduler.RunAfter(say("🎵 Narrator", "[awkward silence]"), 5500*time.Millisecond)
	lipPopping, cancelLipPopping := sayEvery("🎵 Narrator", "*pop*")
	scheduler.RunEveryAfter(lipPopping, 600*time.Millisecond, 6*time.Second)

	// Final explosion
	scheduler.RunAfter(say("👹 Shrek", "THAT'S IT!!!"), 8*time.Second)

	// Stop the lip popping when Fiona announces arrival
	scheduler.RunAfter(func(now time.Time, elapsed time.Duration) bool {
		cancelLipPopping()
		fmt.Println("🎭 [Background sounds stop]")
		return false
	}, 8500*time.Millisecond)

	scheduler.RunAfter(say("👸 Fiona", "We're here!"), 8500*time.Millisecond)
	scheduler.RunAfter(say("🐴 Donkey", "Finally! 🎉"), 9*time.Second)
}
