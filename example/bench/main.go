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
		bench.WithSamples(200),
	)
}

func benchmark(b *bench.B) {
	s := timeline.New()
	var counter atomic.Uint64
	const batch = 100

	emit.On(func(event *ByPointer, now time.Time, elapsed time.Duration) error {
		counter.Add(1)
		return nil
	})

	emit.On(func(event Event, now time.Time, elapsed time.Duration) error {
		counter.Add(1)
		return nil
	})

	b.RunN("task-next", func(i int) int {
		for i := 0; i < batch; i++ {
			s.Run(func(now time.Time, elapsed time.Duration) bool {
				counter.Add(1)
				return true
			})
		}
		s.Tick()
		return batch
	})

	b.RunN("task-after", func(i int) int {
		for i := 0; i < batch; i++ {
			s.RunAfter(func(now time.Time, elapsed time.Duration) bool {
				counter.Add(1)
				return true
			}, time.Duration(10*i)*time.Millisecond)
		}
		s.Tick()
		return batch
	})

	b.RunN("emit-next", func(i int) int {
		for i := 0; i < batch; i++ {
			emit.Next(Event{Number: i, String: "test"})
		}
		return batch
	})

	b.RunN("emit-after", func(i int) int {
		for i := 0; i < batch; i++ {
			emit.After(Event{Number: i, String: "test"}, time.Duration(10*i)*time.Millisecond)
		}
		emit.Default.Tick()
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

type Event struct {
	Number int
	String string
}

func (t Event) Type() uint32 {
	return 0x02
}
