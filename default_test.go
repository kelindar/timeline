package timeline

import (
	"github.com/kelindar/event"
)

/*
func TestSchedule(t *testing.T) {
	now := time.Unix(0, 0)

	Schedule(MyEvent1{Text: "Immediate Event 1"}, now)
	Schedule(MyEvent2{Text: "Immediate Event 2"}, now.Add(5*time.Millisecond))
	Schedule(MyEvent1{Text: "Future Event 1"}, now.Add(495*time.Millisecond))
	Schedule(MyEvent2{Text: "Future Event 2"}, now.Add(1600*time.Millisecond))

	var wg sync.WaitGroup
	wg.Add(4)

	event.On(func(e MyEvent1) { wg.Done() })
	event.On(func(e MyEvent2) { wg.Done() })

	for ts := Tick(0); ts < 200; ts++ {
		Default.Tick(ts)
	}

	wg.Wait()

}*/

// ------------------------------------- Test Events -------------------------------------

const (
	TypeEvent1 = 0x1
	TypeEvent2 = 0x2
)

type MyEvent1 struct{ Text string }

func (t MyEvent1) Type() uint32 { return TypeEvent1 }

func (t MyEvent1) Execute() {
	event.Emit(t)
}

type MyEvent2 struct{ Text string }

func (t MyEvent2) Type() uint32 { return TypeEvent2 }

func (t MyEvent2) Execute() {
	event.Emit(t)
}
