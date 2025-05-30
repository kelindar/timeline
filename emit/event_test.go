package emit

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/*
go test -bench=. -benchmem -benchtime=10s
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkEvent/1x1-24         	13259682	        84.58 ns/op	        11.73 million/s	     169 B/op	       1 allocs/op
BenchmarkEvent/1x10-24        	16216171	       104.8 ns/op	        74.95 million/s	     249 B/op	       1 allocs/op
BenchmarkEvent/1x100-24       	26087012	       669.5 ns/op	        70.51 million/s	     228 B/op	       1 allocs/op
BenchmarkEvent/10x1-24        	 2721086	       510.1 ns/op	        18.33 million/s	     953 B/op	      10 allocs/op
BenchmarkEvent/10x10-24       	 1000000	      1095 ns/op	        50.99 million/s	    2100 B/op	      10 allocs/op
BenchmarkEvent/10x100-24      	 1000000	      1294 ns/op	        57.49 million/s	    2151 B/op	      10 allocs/op
*/
func BenchmarkEvent(b *testing.B) {
	for _, topics := range []int{1, 10} {
		for _, subs := range []int{1, 10, 100} {
			b.Run(fmt.Sprintf("%dx%d", topics, subs), func(b *testing.B) {
				var count atomic.Int64
				for i := 0; i < subs; i++ {
					for id := 10; id < 10+topics; id++ {
						defer OnType(uint32(id), func(ev Dynamic, now time.Time, elapsed time.Duration) error {
							count.Add(1)
							return nil
						})()
					}
				}

				b.ReportAllocs()
				b.ResetTimer()

				start := time.Now()
				for n := 0; n < b.N; n++ {
					for id := 10; id < 10+topics; id++ {
						Next(Dynamic{ID: id})
					}
				}

				elapsed := time.Since(start)
				rate := float64(count.Load()) / 1e6 / elapsed.Seconds()
				b.ReportMetric(rate, "million/s")
			})
		}
	}
}

func TestEmit(t *testing.T) {
	events := make(chan MyEvent2)
	defer On(func(ev MyEvent2, now time.Time, elapsed time.Duration) error {
		assert.Equal(t, "Hello", ev.Text)
		events <- ev
		return nil
	})()

	// Emit the event
	Next(MyEvent2{Text: "Hello"})
	<-events

	At(MyEvent2{Text: "Hello"}, time.Now().Add(40*time.Millisecond))
	<-events

	After(MyEvent2{Text: "Hello"}, 20*time.Millisecond)
	<-events

	EveryAt(MyEvent2{Text: "Hello"}, 50*time.Millisecond, time.Now().Add(10*time.Millisecond))
	<-events

	EveryAfter(MyEvent2{Text: "Hello"}, 30*time.Millisecond, 10*time.Millisecond)
	<-events

	Every(MyEvent2{Text: "Hello"}, 10*time.Millisecond)
	<-events
}

func TestOnType(t *testing.T) {
	events := make(chan Dynamic)
	defer OnType(42, func(ev Dynamic, now time.Time, elapsed time.Duration) error {
		assert.Equal(t, 42, ev.ID)
		events <- ev
		return nil
	})()

	// Emit the event
	Next(Dynamic{ID: 42})
	<-events
}

func TestOnError(t *testing.T) {
	errors := make(chan error)
	defer OnError(func(err error, about any) {
		errors <- err
	})()

	defer On(func(ev MyEvent2, now time.Time, elapsed time.Duration) error {
		return fmt.Errorf("On()")
	})()

	// Emit the event
	Error(fmt.Errorf("Err"), nil)
	assert.Equal(t, "Err", (<-errors).Error())

	// Fail in the handler
	Next(MyEvent2{})
	assert.Equal(t, "On()", (<-errors).Error())

}

func TestOnTypeError(t *testing.T) {
	errors := make(chan error)
	defer OnError(func(err error, about any) {
		errors <- err
	})()

	defer OnType(42, func(ev Dynamic, now time.Time, elapsed time.Duration) error {
		return fmt.Errorf("OnType()")
	})()

	// Fail in dynamic event handler
	Next(Dynamic{ID: 42})
	assert.Equal(t, "OnType()", (<-errors).Error())
}

func TestOnEvery(t *testing.T) {
	events := make(chan MyEvent2)
	defer OnEvery(func(now time.Time, elapsed time.Duration) error {
		events <- MyEvent2{}
		return nil
	}, 20*time.Millisecond)()

	// Emit the event
	<-events
	<-events
	<-events
}

func TestEveryCancel(t *testing.T) {
	events := make(chan MyEvent2, 10)
	defer On(func(ev MyEvent2, now time.Time, elapsed time.Duration) error {
		events <- ev
		return nil
	})()

	// Start recurring event
	cancel := Every(MyEvent2{Text: "Recurring"}, 20*time.Millisecond)

	// Wait for a few events
	<-events
	<-events
	<-events

	// Cancel the recurring event
	cancel()

	// Wait a bit to ensure no more events come
	time.Sleep(100 * time.Millisecond)

	// Channel should not have more events (non-blocking check)
	select {
	case <-events:
		t.Error("Expected no more events after cancellation")
	default:
		// Good, no more events
	}
}

// ------------------------------------- Test Events -------------------------------------

const (
	TypeEvent1 = 0x1
	TypeEvent2 = 0x2
)

type MyEvent1 struct {
	Number int
}

func (t MyEvent1) Type() uint32 { return TypeEvent1 }

type MyEvent2 struct {
	Text string
}

func (t MyEvent2) Type() uint32 { return TypeEvent2 }

type Dynamic struct {
	ID int
}

func (t Dynamic) Type() uint32 {
	return uint32(t.ID)
}
