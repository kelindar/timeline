// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package emit

import (
	"context"
	"math"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/kelindar/event"
	"github.com/kelindar/timeline"
)

const (
	resolution = 10 * time.Millisecond
	buckets    = int(time.Second / resolution)
)

// Scheduler is the default scheduler used to emit events.
var Scheduler = func() *timeline.Scheduler {
	s := timeline.New()
	s.Start(context.Background())
	s.RunEvery(func(now time.Time, elapsed time.Duration) bool {
		idx := int(driverIdx.Add(1)-1) % buckets
		wheels.Range(func(_ any, v any) bool {
			v.(flushFunc)(idx, now, elapsed)
			return true
		})
		return true
	}, resolution)
	return s
}()

// ----------------------------------------- Driver (time wheel) -----------------------------------------

var (
	driverIdx   atomic.Int32
	wheels      sync.Map // map[uint32]flushFunc
	typedWheels sync.Map // map[uint32]any (*wheel[T])
)

type flushFunc = func(idx int, now time.Time, elapsed time.Duration)

type wheel[T event.Event] struct {
	mu      sync.Mutex
	current uint32 // Current tick position for this wheel
	buckets [][]T
}

func newTypedWheel[T event.Event]() *wheel[T] {
	w := &wheel[T]{
		buckets: make([][]T, buckets),
	}
	for i := 0; i < buckets; i++ {
		w.buckets[i] = make([]T, 0, 16)
	}
	return w
}

func wheelOf[T event.Event]() *wheel[T] {
	key := hashOfT[T]()
	if v, ok := typedWheels.Load(key); ok {
		return v.(*wheel[T])
	}

	w := newTypedWheel[T]()
	actual, loaded := typedWheels.LoadOrStore(key, w)
	tw := actual.(*wheel[T])
	if !loaded {
		wheels.Store(key, tw.flush)
	}
	return tw
}

func (w *wheel[T]) flush(idx int, now time.Time, elapsed time.Duration) {
	w.mu.Lock()
	w.current = uint32(idx) // Update wheel's current position
	batch := w.buckets[idx]
	w.buckets[idx] = w.buckets[idx][:0]
	w.mu.Unlock()
	for i := range batch {
		event.Publish(event.Default, signal[T]{
			Data:    batch[i],
			Time:    now,
			Elapsed: elapsed,
		})
	}
}

func emit[T event.Event](ev T, delta int) {
	w := wheelOf[T]()
	w.mu.Lock()
	idx := (int(w.current) + delta) % buckets
	w.buckets[idx] = append(w.buckets[idx], ev)
	w.mu.Unlock()
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
	emit(ev, 1)
}

// At writes an event at specific 'at' time.
func At[T event.Event](ev T, at time.Time) {
	delta := int(time.Until(at) / resolution)
	if delta < 1 {
		delta = 1
	}
	emit(ev, delta)
}

// After writes an event after a 'delay'.
func After[T event.Event](ev T, after time.Duration) {
	steps := int(after / resolution)
	if steps < 1 {
		steps = 1
	}
	emit(ev, steps)
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

func hashOfT[T any]() uint32 {
	var result T
	return loadHash(reflect.TypeOf(result))
}

// loadHash loads the hash of the given type, this is a hack to avoid
// time consuming hashing every time and is not guaranteed to work in
// future versions of Go.
func loadHash(rt reflect.Type) uint32 {
	return (*rtype)(unsafe.Pointer((*iface)(unsafe.Pointer(&rt)).data)).hash
}

type rtype struct {
	size    uintptr
	ptrdata uintptr // number of bytes in the type that can contain pointers
	hash    uint32  // this is the unexported field
	// ... rest omitted
}

// This struct matches the memory layout of an interface in Go.
type iface struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}
