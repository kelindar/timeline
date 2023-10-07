// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package event

import (
	"context"
	"time"

	"github.com/kelindar/event"
	"github.com/kelindar/timeline"
)

// Scheduler is the default scheduler used to emit events.
var Scheduler = func() *timeline.Scheduler {
	s := timeline.New()
	s.Start(context.Background())
	return s
}()

// signal represents a signal event
type Event[T event.Event] struct {
	Time    time.Time     // The time at which the event was emitted
	Elapsed time.Duration // The time elapsed since the last event
	Data    T
}

// Type returns the type of the event
func (e Event[T]) Type() uint32 {
	return e.Data.Type()
}

// emit writes an event into the dispatcher
func emit[T event.Event](ev T) func(now time.Time, elapsed time.Duration) bool {
	return func(now time.Time, elapsed time.Duration) bool {
		event.Publish(event.Default, Event[T]{
			Data:    ev,
			Time:    now,
			Elapsed: elapsed,
		})
		return true
	}
}

// On subscribes to an event, the type of the event will be automatically
// inferred from the provided type. Must be constant for this to work. This
// functions same way as Subscribe() but uses the default dispatcher instead.
func On[T event.Event](handler func(Event[T])) context.CancelFunc {
	return event.Subscribe[Event[T]](event.Default, handler)
}

// OnType subscribes to an event with the specified event type. This functions
// same way as SubscribeTo() but uses the default dispatcher instead.
func OnType[T event.Event](eventType uint32, handler func(Event[T])) context.CancelFunc {
	return event.SubscribeTo[Event[T]](event.Default, eventType, handler)
}

// Emit writes an event during the next tick.
func Emit[T event.Event](ev T) {
	Scheduler.Run(emit(ev))
}

// EmitAt writes an event at specific 'at' time.
func EmitAt[T event.Event](ev T, at time.Time) {
	Scheduler.RunAt(emit(ev), at)
}

// EmitAfter writes an event after a 'delay'.
func EmitAfter[T event.Event](ev T, after time.Duration) {
	Scheduler.RunAfter(emit(ev), after)
}

// EmitEvery writes an event at 'interval' intervals, starting at the next boundary tick.
func EmitEvery[T event.Event](ev T, interval time.Duration) {
	Scheduler.RunEvery(emit(ev), interval)
}

// EmitEveryAt writes an event at 'interval' intervals, starting at 'startTime'.
func EmitEveryAt[T event.Event](ev T, interval time.Duration, startTime time.Time) {
	Scheduler.RunEveryAt(emit(ev), interval, startTime)
}

// EmitEveryAfter writes an event at 'interval' intervals after a 'delay'.
func EmitEveryAfter[T event.Event](ev T, interval time.Duration, delay time.Duration) {
	Scheduler.RunEveryAfter(emit(ev), interval, delay)
}
