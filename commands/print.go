package commands

import (
	"fmt"
	"time"
)

func (b *Boom) Print() {
	total := b.end.Sub(b.start)
	totalSuccessful := 0
	var avgTotal int64
	var fastest, slowest time.Duration

	for {
		select {
		case r := <-b.results:
			if !r.is2xx {
				continue
			}
			avgTotal += r.dur.Nanoseconds()
			if fastest.Nanoseconds() == 0 || r.dur.Nanoseconds() < fastest.Nanoseconds() {
				fastest = r.dur
			}
			if r.dur.Nanoseconds() > slowest.Nanoseconds() {
				slowest = r.dur
			}
			totalSuccessful++
		default:
			rps := float64(b.N) / total.Seconds()
			fmt.Printf("\n")
			fmt.Printf("Summary:\n")
			fmt.Printf("  total:\t%v requests\n", b.N)
			fmt.Printf("  concurrency:\t%v concurrent requests\n", b.C)
			fmt.Printf("  total 2xx:\t%v requests\n", totalSuccessful)
			fmt.Printf("  total:\t%v secs\n", total.Seconds())
			fmt.Printf("  slowest:\t%v secs\n", slowest.Seconds())
			fmt.Printf("  fastest:\t%v secs\n", fastest.Seconds())
			fmt.Printf("  average:\t%v nanoseconds\n", avgTotal/int64(b.N)) // TODO: in seconds
			fmt.Printf("  requests/sec:\t%v\n", rps)
			fmt.Printf("  speed index:\t%v\n", speedIndex(rps))
			return
		}
	}
}

func speedIndex(rps float64) string {
	if rps > 500 {
		return "Whoa, pretty neat"
	} else if rps > 100 {
		return "Pretty good"
	} else if rps > 50 {
		return "Meh"
	} else {
		return "Hahahaha"
	}
}
