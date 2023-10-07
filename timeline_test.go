package timeline

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

var counter atomic.Uint64

/*
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkEvent/batch/1-24         	34783515	        34.82 ns/op	        34.78 million/op	       0 B/op	       0 allocs/op
BenchmarkEvent/batch/10-24        	 5811174	       209.1 ns/op	        58.11 million/op	       0 B/op	       0 allocs/op
BenchmarkEvent/batch/100-24       	  600006	      1988 ns/op	        60.00 million/op	       0 B/op	       0 allocs/op
BenchmarkEvent/batch/1000-24      	   43273	     26310 ns/op	        43.27 million/op	       0 B/op	       0 allocs/op
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
					//Schedule(CountEvent{}, when)
					Default.RunAt(func() bool {
						counter.Add(1)
						return true
					}, when)
				}

				Default.Tick(Tick(n))
			}

			b.ReportMetric(float64(counter.Load())/1000000, "million/op")
		})
	}
}
