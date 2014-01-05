package commands

import (
	"fmt"
	"math"
	"time"
)

var statusCodeDist map[int]int = make(map[int]int)

func (b *Boom) Print() {
	total := b.end.Sub(b.start)
	var avgTotal int64
	var fastest, slowest time.Duration

	for {
		select {
		case r := <-b.results:
			statusCodeDist[r.statusCode]++

			avgTotal += r.dur.Nanoseconds()
			if fastest.Nanoseconds() == 0 || r.dur.Nanoseconds() < fastest.Nanoseconds() {
				fastest = r.dur
			}
			if r.dur.Nanoseconds() > slowest.Nanoseconds() {
				slowest = r.dur
			}
		default:
			rps := float64(b.N) / total.Seconds()
			fmt.Printf("\nSummary:\n")
			fmt.Printf("  total:\t%v secs\n", total.Seconds())
			fmt.Printf("  slowest:\t%v secs\n", slowest.Seconds())
			fmt.Printf("  fastest:\t%v secs\n", fastest.Seconds())
			fmt.Printf("  average:\t%v secs\n", float64(avgTotal)/float64(b.N)*math.Pow(10, 9)) // TODO: in seconds
			fmt.Printf("  requests/sec:\t%v\n", rps)
			fmt.Printf("  speed index:\t%v\n", speedIndex(rps))
			b.printStatusCodes()
			return
		}
	}
}

func (b *Boom) printStatusCodes() {
	fmt.Printf("\nStatus code distrubution:\n")
	for code, num := range statusCodeDist {
		fmt.Printf("  [%d]\t%d responses\n", code, num)
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
