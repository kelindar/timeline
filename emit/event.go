// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package emit

import (
	"context"
	"math"
	"sync/atomic"
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

// ----------------------------------------- Forward Event -----------------------------------------

// signal represents a forwarded event
type signal[T event.Event] struct {
	Time    time.Time     // The time at which the event was emitted
	Elapsed time.Duration // The time elapsed since the last event
	Data    T
}

// Type returns the type of the event
func (e signal[T]) Type() uint32 {
	return e.Data.Type()
}

// ----------------------------------------- Error Event -----------------------------------------

// fault represents an error event
type fault struct {
	error
	About any // The context of the error
}

// Type returns the type of the event
func (e fault) Type() uint32 {
	return math.MaxUint32
}

// ----------------------------------------- Timer Event -----------------------------------------

var nextTimerID uint32 = 1 << 30

// Timer represents a Timer event
type Timer struct {
	ID uint32
}

// Type returns the type of the event
func (e Timer) Type() uint32 {
	return e.ID
}

// ----------------------------------------- Subscribe -----------------------------------------

// On subscribes to an event, the type of the event will be automatically
// inferred from the provided type. Must be constant for this to work.
func On[T event.Event](handler func(event T, now time.Time, elapsed time.Duration) error) context.CancelFunc {
	return event.Subscribe[signal[T]](event.Default, func(m signal[T]) {
		if err := handler(m.Data, m.Time, m.Elapsed); err != nil {
			Error(err, m.Data)
		}
	})
}

// OnType subscribes to an event with the specified event type.
func OnType[T event.Event](eventType uint32, handler func(event T, now time.Time, elapsed time.Duration) error) context.CancelFunc {
	return event.SubscribeTo[signal[T]](event.Default, eventType, func(m signal[T]) {
		if err := handler(m.Data, m.Time, m.Elapsed); err != nil {
			Error(err, m.Data)
		}
	})
}

// OnError subscribes to an error event.
func OnError(handler func(err error, about any)) context.CancelFunc {
	return event.Subscribe[fault](event.Default, func(m fault) {
		handler(m.error, m.About)
	})
}

// OnEvery creates a timer that fires every 'interval' and calls the handler.
func OnEvery(handler func(now time.Time, elapsed time.Duration) error, interval time.Duration) context.CancelFunc {
	id := atomic.AddUint32(&nextTimerID, 1)
	if id >= (math.MaxUint32 - 1) {
		panic("emit: too many timers created")
	}

	// Subscribe to the timer event
	cancel := OnType[Timer](id, func(_ Timer, now time.Time, elapsed time.Duration) error {
		return handler(now, elapsed)
	})

	// Start the timer
	Every(Timer{ID: id}, interval)
	return cancel
}

// ----------------------------------------- Publish -----------------------------------------

// Next writes an event during the next tick.
func Next[T event.Event](ev T) {
	Scheduler.Run(emit(ev))
}

// At writes an event at specific 'at' time.
func At[T event.Event](ev T, at time.Time) {
	Scheduler.RunAt(emit(ev), at)
}

// After writes an event after a 'delay'.
func After[T event.Event](ev T, after time.Duration) {
	Scheduler.RunAfter(emit(ev), after)
}

// Every writes an event at 'interval' intervals, starting at the next boundary tick.
// Returns a cancel function to stop the recurring event.
func Every[T event.Event](ev T, interval time.Duration) context.CancelFunc {
	return emitEvery(ev, interval, func(task timeline.Task, interval time.Duration) {
		Scheduler.RunEvery(task, interval)
	})
}

// EveryAt writes an event at 'interval' intervals, starting at 'startTime'.
// Returns a cancel function to stop the recurring event.
func EveryAt[T event.Event](ev T, interval time.Duration, startTime time.Time) context.CancelFunc {
	return emitEvery(ev, interval, func(task timeline.Task, interval time.Duration) {
		Scheduler.RunEveryAt(task, interval, startTime)
	})
}

// EveryAfter writes an event at 'interval' intervals after a 'delay'.
// Returns a cancel function to stop the recurring event.
func EveryAfter[T event.Event](ev T, interval time.Duration, delay time.Duration) context.CancelFunc {
	return emitEvery(ev, interval, func(task timeline.Task, interval time.Duration) {
		Scheduler.RunEveryAfter(task, interval, delay)
	})
}

// Error writes an error event.
func Error(err error, about any) {
	event.Publish(event.Default, fault{
		error: err,
		About: about,
	})
}

// emit writes an event into the dispatcher
func emit[T event.Event](ev T) func(now time.Time, elapsed time.Duration) bool {
	return func(now time.Time, elapsed time.Duration) bool {
		event.Publish(event.Default, signal[T]{
			Data:    ev,
			Time:    now,
			Elapsed: elapsed,
		})
		return true
	}
}

// emitEvery creates a cancellable recurring event
func emitEvery[T event.Event](ev T, interval time.Duration, scheduler func(timeline.Task, time.Duration)) func() {
	var cancelled atomic.Bool
	task := func(now time.Time, elapsed time.Duration) bool {
		event.Publish(event.Default, signal[T]{
			Data:    ev,
			Time:    now,
			Elapsed: elapsed,
		})
		return !cancelled.Load()
	}

	scheduler(task, interval)
	return func() {
		cancelled.Store(true)
	}
}
