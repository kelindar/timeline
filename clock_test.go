package timeline

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTickOf(t *testing.T) {
	tc := map[Tick]time.Duration{
		0:      0,
		1:      10 * time.Millisecond,
		2:      20 * time.Millisecond,
		10:     100 * time.Millisecond,
		100:    time.Second,
		101:    time.Second + 10*time.Millisecond,
		360000: time.Hour,
	}

	for expect, duration := range tc {
		assert.Equal(t, expect, TickOf(time.Unix(0, int64(duration))))
	}
}
