package main

import (
	"sync/atomic"
	"time"

	"github.com/kelindar/bench"
	"github.com/kelindar/timeline"
	"github.com/kelindar/timeline/emit"
)

func main() {
	bench.Run(benchmark,
		bench.WithDuration(5*time.Millisecond),
		bench.WithSamples(100),
	)
}

func benchmark(b *bench.B) {
	s := timeline.New()
	var counter atomic.Uint64
	const batch = 10

	emit.On(func(event *ByPointer, now time.Time, elapsed time.Duration) error {
		counter.Add(1)
		return nil
	})

	emit.On(func(event ByValue, now time.Time, elapsed time.Duration) error {
		counter.Add(1)
		return nil
	})

	b.RunN("timeline-next", func(i int) int {
		for i := 0; i < batch; i++ {
			s.Run(func(now time.Time, elapsed time.Duration) bool {
				counter.Add(1)
				return true
			})
		}
		s.Tick()
		return batch
	})

	b.RunN("timeline-after", func(i int) int {
		for i := 0; i < batch; i++ {
			s.RunAfter(func(now time.Time, elapsed time.Duration) bool {
				counter.Add(1)
				return true
			}, time.Duration(10*i)*time.Millisecond)
		}
		s.Tick()
		return batch
	})

	b.RunN("emit-next-ptr", func(i int) int {
		for i := 0; i < batch; i++ {
			emit.Next(&ByPointer{Number: i, String: "test"})
		}
		s.Tick()
		return batch
	})

	b.RunN("emit-after-ptr", func(i int) int {
		for i := 0; i < batch; i++ {
			emit.After(&ByPointer{Number: i, String: "test"}, time.Duration(10*i)*time.Millisecond)
		}
		s.Tick()
		return batch
	})

	b.RunN("emit-next-val", func(i int) int {
		for i := 0; i < batch; i++ {
			emit.Next(ByValue{Number: i, String: "test"})
		}
		s.Tick()
		return batch
	})

	b.RunN("emit-after-val", func(i int) int {
		for i := 0; i < batch; i++ {
			emit.After(ByValue{Number: i, String: "test"}, time.Duration(10*i)*time.Millisecond)
		}
		s.Tick()
		return batch
	})

}

type ByPointer struct {
	Number int
	String string
}

func (t *ByPointer) Type() uint32 {
	return 0x01
}

type ByValue struct {
	Number int
	String string
}

func (t ByValue) Type() uint32 {
	return 0x02
}
