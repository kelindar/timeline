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
BenchmarkEvent/batch/1-24         	33334166	        35.35 ns/op	        33.33 million/op	       0 B/op	       0 allocs/op
BenchmarkEvent/batch/10-24        	 5529952	       220.0 ns/op	        55.30 million/op	       0 B/op	       0 allocs/op
BenchmarkEvent/batch/100-24       	  558123	      2044 ns/op	        55.81 million/op	       0 B/op	       0 allocs/op
BenchmarkEvent/batch/1000-24      	   35406	     31209 ns/op	        35.41 million/op	       0 B/op	       0 allocs/op
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
					Default.RunAt(func() {
						counter.Add(1)
					}, when)
				}

				Default.Tick(Tick(n))
			}

			b.ReportMetric(float64(counter.Load())/1000000, "million/op")
		})
	}
}
