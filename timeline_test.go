package timeline

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

/*
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkEvent/batch/1-24         	35294739	        34.28 ns/op	        35.29 million/op	       0 B/op	       0 allocs/op
BenchmarkEvent/batch/10-24        	 5882352	       206.5 ns/op	        58.82 million/op	       0 B/op	       0 allocs/op
BenchmarkEvent/batch/100-24       	  600008	      1942 ns/op	        60.00 million/op	       0 B/op	       0 allocs/op
BenchmarkEvent/batch/1000-24      	   44310	     25773 ns/op	        44.31 million/op	       0 B/op	       0 allocs/op
*/
func BenchmarkEvent(b *testing.B) {
	for _, size := range []int{1, 10, 100, 1000} {
		b.Run(fmt.Sprintf("batch/%d", size), func(b *testing.B) {
			counter.Store(0)
			now := time.Unix(0, 0)

			b.ReportAllocs()
			b.ResetTimer()

			for n := 0; n < b.N; n++ {
				for i := 0; i < size; i++ {
					when := now.Add(time.Duration(100*i) * time.Millisecond)
					Schedule(CountEvent{}, when)
				}

				defaultTimeline.Tick(Tick(n))
			}

			b.ReportMetric(float64(counter.Load())/1000000, "million/op")
		})
	}
}

var counter atomic.Uint64

type CountEvent struct {
}

func (t CountEvent) Execute() {
	counter.Add(1)
}
