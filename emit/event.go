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
	q := actual.(*queue[T])
	if !loaded {
		s.RunEvery(q.Drain, resolution)
	}
	return q
}

// ----------------------------------------- Subscribe -----------------------------------------

// On subscribes to an event, the type of the event will be automatically
// inferred from the provided type. Must be constant for this to work.
func On[T event.Event](handler func(event T, now time.Time, elapsed time.Duration) error) context.CancelFunc {
	return event.Subscribe(event.Default, func(m signal[T]) {
		if err := handler(m.Data, m.Time.Time(), m.Elapsed.Duration()); err != nil {
			Error(err, m.Data)
		}
	})
}

// OnType subscribes to an event with the specified event type.
func OnType[T event.Event](eventType uint32, handler func(event T, now time.Time, elapsed time.Duration) error) context.CancelFunc {
	return event.SubscribeTo(event.Default, eventType, func(m signal[T]) {
		if err := handler(m.Data, m.Time.Time(), m.Elapsed.Duration()); err != nil {
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
	var cancelled atomic.Bool
	Default.RunEvery(func(now time.Time, elapsed time.Duration) bool {
		if err := handler(now, elapsed); err != nil {
			Error(err, nil)
		}

		return !cancelled.Load()
	}, interval)

	return func() {
		cancelled.Store(true)
	}
}

// ----------------------------------------- Publish -----------------------------------------

// Next writes an event during the next tick. Zero allocation.
func Next[T event.Event](ev T) {
	q := queueOf[T](Default, ev.Type())
	q.Push(ev)
}

// After writes an event after the specified delay. This one allocates a closure.
func After[T event.Event](ev T, delay time.Duration) {
	Default.RunAfter(func(now time.Time, elapsed time.Duration) bool {
		event.Publish(event.Default, signal[T]{
			Data:    ev,
			Time:    tickOf(now),
			Elapsed: durationOf(elapsed),
		})
		return true
	}, delay)
}

// Error writes an error event.
func Error(err error, about any) {
	event.Publish(event.Default, fault{
		error: err,
		About: about,
	})
}

// ----------------------------------------- Forward Event -----------------------------------------

// signal represents a forwarded event
type signal[T event.Event] struct {
	Time    tick // The time at which the event was emitted
	Elapsed span // The time elapsed since the last event
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

// ----------------------------------------- Time (in ticks) -----------------------------------------

// tick represents a point in time, rounded up to the resolution of the clock.
type tick int64

// Time converts the tick to a timestamp.
func (t tick) Time() time.Time {
	return time.Unix(0, int64(t)*int64(resolution))
}

// tickOf returns the time rounded up to the resolution of the clock.
func tickOf(t time.Time) tick {
	return tick(t.UnixNano() / int64(resolution))
}

// ----------------------------------------- Duration (in ticks) -----------------------------------------

// span represents a time span (duration) in ticks
type span uint32

// Duration converts the span to a duration.
func (s span) Duration() time.Duration {
	return time.Duration(s) * resolution
}

// durationOf computes a duration in terms of ticks.
func durationOf(t time.Duration) span {
	return span(t / resolution)
}
