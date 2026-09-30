package requester

import (
	"testing"
	"time"
)

func TestQPCDurationAfterLongUptime(t *testing.T) {
	const frequency = int64(10_000_000)
	for _, want := range []time.Duration{15 * time.Minute, time.Hour, 24*time.Hour + 500*time.Millisecond} {
		t.Run(want.String(), func(t *testing.T) {
			ticks := int64(want/time.Second)*frequency + int64(want%time.Second)*frequency/int64(time.Second)
			if got := qpcDuration(ticks, frequency); got != want {
				t.Fatalf("qpcDuration(%d, %d) = %s, want %s", ticks, frequency, got, want)
			}
		})
	}
}
