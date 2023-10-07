package event

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/*
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkEvent/1x1-24         	17777803	        60.10 ns/op	        16.45 million/s	     170 B/op	       1 allocs/op
BenchmarkEvent/1x10-24        	16062572	       162.3 ns/op	        67.38 million/s	     220 B/op	       1 allocs/op
BenchmarkEvent/1x100-24       	22221769	       106.9 ns/op	        56.41 million/s	     205 B/op	       1 allocs/op
BenchmarkEvent/10x1-24        	 3183024	       459.8 ns/op	        23.68 million/s	     920 B/op	      10 allocs/op
BenchmarkEvent/10x10-24       	 1465628	      1136 ns/op	        55.92 million/s	    1981 B/op	      10 allocs/op
BenchmarkEvent/10x100-24      	 1432830	      1780 ns/op	        61.69 million/s	    2130 B/op	      10 allocs/op
*/
func BenchmarkEvent(b *testing.B) {
	for _, topics := range []int{1, 10} {
		for _, subs := range []int{1, 10, 100} {
			b.Run(fmt.Sprintf("%dx%d", topics, subs), func(b *testing.B) {
				var count atomic.Int64
				for i := 0; i < subs; i++ {
					for id := 10; id < 10+topics; id++ {
						defer OnType(uint32(id), func(ev Event[Dynamic]) {
							count.Add(1)
						})()
					}
				}

				start := time.Now()
				b.ReportAllocs()
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					for id := 10; id < 10+topics; id++ {
						Emit(Dynamic{ID: id})
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
	defer On(func(ev Event[MyEvent2]) {
		assert.Equal(t, "Hello", ev.Data.Text)
		events <- ev.Data
	})()

	// Emit the event
	Emit(MyEvent2{Text: "Hello"})
	<-events

	EmitAt(MyEvent2{Text: "Hello"}, time.Now().Add(40*time.Millisecond))
	<-events

	EmitAfter(MyEvent2{Text: "Hello"}, 20*time.Millisecond)
	<-events

	EmitEveryAt(MyEvent2{Text: "Hello"}, 50*time.Millisecond, time.Now().Add(10*time.Millisecond))
	<-events

	EmitEveryAfter(MyEvent2{Text: "Hello"}, 30*time.Millisecond, 10*time.Millisecond)
	<-events

	EmitEvery(MyEvent2{Text: "Hello"}, 10*time.Millisecond)
	<-events
}

func TestOnType(t *testing.T) {
	events := make(chan Dynamic)
	defer OnType(42, func(ev Event[Dynamic]) {
		assert.Equal(t, 42, ev.Data.ID)
		events <- ev.Data
	})()

	// Emit the event
	Emit(Dynamic{ID: 42})
	<-events
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
