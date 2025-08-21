// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package emit

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kelindar/event"
	"github.com/kelindar/timeline"
)

const (
	resolution = 10 * time.Millisecond
)

// Scheduler is the default scheduler used to emit events on a timeline
type Scheduler struct {
	*timeline.Scheduler
	queues sync.Map
}

// Default is the default scheduler used to emit events.
var Default = func() *Scheduler {
	s := &Scheduler{
		Scheduler: timeline.New(),
	}

	s.Start(context.Background())
	return s
}()

// ----------------------------------------- Queues -----------------------------------------

func queueOf[T event.Event](s *Scheduler, eventType uint32) *queue[T] {
	if v, ok := s.queues.Load(eventType); ok {
		return v.(*queue[T])
	}

	actual, loaded := s.queues.LoadOrStore(eventType, newQueue[T]())
	w := actual.(*queue[T])
	if !loaded {
		s.RunEvery(w.Drain, resolution)
	}
	return w
}

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
	return event.Subscribe(event.Default, func(m signal[T]) {
		if err := handler(m.Data, m.Time, m.Elapsed); err != nil {
			Error(err, m.Data)
		}
	})
}

// OnType subscribes to an event with the specified event type.
func OnType[T event.Event](eventType uint32, handler func(event T, now time.Time, elapsed time.Duration) error) context.CancelFunc {
	return event.SubscribeTo(event.Default, eventType, func(m signal[T]) {
		if err := handler(m.Data, m.Time, m.Elapsed); err != nil {
			Error(err, m.Data)
		}
	})
}

// OnError subscribes to an error event.
func OnError(handler func(err error, about any)) context.CancelFunc {
	return event.Subscribe(event.Default, func(m fault) {
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
	cancel := OnType(id, func(_ Timer, now time.Time, elapsed time.Duration) error {
		return handler(now, elapsed)
	})

	// Start the timer
	Every(Timer{ID: id}, interval)
	return cancel
}

// ----------------------------------------- Publish -----------------------------------------

// Next writes an event during the next tick.
func Next[T event.Event](ev T) {
	w := queueOf[T](Default, ev.Type())
	w.Push(ev)
}

// Every writes an event at 'interval' intervals, starting at the next boundary tick.
// Returns a cancel function to stop the recurring event.
func Every[T event.Event](ev T, interval time.Duration) context.CancelFunc {
	return emitEvery(ev, interval, func(task timeline.Task, interval time.Duration) {
		Default.RunEvery(task, interval)
	})
}

// Error writes an error event.
func Error(err error, about any) {
	event.Publish(event.Default, fault{
		error: err,
		About: about,
	})
}

// emitEvery creates a cancellable recurring event with optimized closure
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
