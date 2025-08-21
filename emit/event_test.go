package emit

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/*
go test -bench=. -benchmem -benchtime=10s
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkEvent/1x1-24         	49567932	        23.05 ns/op	        29.03 million/s	       3 B/op	       0 allocs/op
BenchmarkEvent/1x10-24        	57662643	        23.18 ns/op	        60.29 million/s	       8 B/op	       0 allocs/op
BenchmarkEvent/1x100-24       	60567517	        54.51 ns/op	        72.32 million/s	      10 B/op	       0 allocs/op
BenchmarkEvent/10x1-24        	 5801366	       224.5 ns/op	        49.58 million/s	      86 B/op	       0 allocs/op
BenchmarkEvent/10x10-24       	 5055789	       218.1 ns/op	        81.23 million/s	     108 B/op	       0 allocs/op
BenchmarkEvent/10x100-24      	 2404316	       512.2 ns/op	        71.51 million/s	     125 B/op	       0 allocs/op
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

	After(MyEvent2{Text: "Hello"}, 20*time.Millisecond)
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

func TestOnEveryCancel(t *testing.T) {
	var count atomic.Int32
	cancel := OnEvery(func(now time.Time, elapsed time.Duration) error {
		count.Add(1)
		return nil
	}, 10*time.Millisecond)

	cancel()

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, int(count.Load()))
}

func TestStress(t *testing.T) {
	const count = 1000000

	var wg sync.WaitGroup
	wg.Add(count)
	defer OnType(1234, func(ev Dynamic, now time.Time, elapsed time.Duration) error {
		wg.Done()
		return nil
	})()

	for i := 0; i < count; i++ {
		Next(Dynamic{ID: 1234})
	}

	wg.Wait()
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
