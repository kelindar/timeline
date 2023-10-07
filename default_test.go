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
BenchmarkEvent/batch/1-24         	32432432	        37.19 ns/op	        32.43 million/op	       0 B/op	       0 allocs/op
BenchmarkEvent/batch/10-24        	 6434312	       187.0 ns/op	        64.34 million/op	       0 B/op	       0 allocs/op
BenchmarkEvent/batch/100-24       	  571430	      2041 ns/op	        57.09 million/op	       7 B/op	       0 allocs/op
BenchmarkEvent/batch/1000-24      	   13340	     92729 ns/op	         8.345 million/op	   51714 B/op	       0 allocs/op
*/
func BenchmarkEvent(b *testing.B) {
	for _, size := range []int{1, 10, 100, 1000} {
		b.Run(fmt.Sprintf("batch/%d", size), func(b *testing.B) {

			counter.Store(0)
			s := New()

			b.ReportAllocs()
			b.ResetTimer()

			for n := 0; n < b.N; n++ {
				for i := 0; i < size; i++ {
					s.RunAfter(func(time.Time) bool {
						counter.Add(1)
						return true
					}, time.Duration(100*i)*time.Millisecond)
				}

				s.Tick()
			}

			b.ReportMetric(float64(counter.Load())/1000000, "million/op")
		})
	}
}
