// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

var statusCodeDist map[int]int = make(map[int]int)

var latencies []float64

func (b *Boom) Print() {
	total := b.end.Sub(b.start)
	var avgTotal float64
	var fastest, slowest time.Duration

	for {
		select {
		case r := <-b.results:
			latencies = append(latencies, r.duration.Seconds())
			statusCodeDist[r.statusCode]++

			avgTotal += r.duration.Seconds()
			if fastest.Nanoseconds() == 0 || r.duration.Nanoseconds() < fastest.Nanoseconds() {
				fastest = r.duration
			}
			if r.duration.Nanoseconds() > slowest.Nanoseconds() {
				slowest = r.duration
			}
		default:
			rps := float64(b.N) / total.Seconds()
			fmt.Printf("\nSummary:\n")
			fmt.Printf("  Total:\t%4.4f secs.\n", total.Seconds())
			fmt.Printf("  Slowest:\t%4.4f secs.\n", slowest.Seconds())
			fmt.Printf("  Fastest:\t%4.4f secs.\n", fastest.Seconds())
			fmt.Printf("  Average:\t%4.4f secs.\n", avgTotal/float64(b.N))
			fmt.Printf("  Requests/sec:\t%4.4f\n", rps)
			fmt.Printf("  Speed index:\t%v\n", speedIndex(rps))
			sort.Float64s(latencies)
			b.printHistogram()
			b.printLatencies()
			b.printStatusCodes()
			return
		}
	}
}

// Prints percentile latencies.
func (b *Boom) printLatencies() {
	pctls := []int{10, 25, 50, 75, 90, 95, 99}
	// Sort the array
	data := make([]float64, len(pctls))
	j := 0
	for i := 0; i < len(latencies) && j < len(pctls); i++ {
		current := (i + 1) * 100 / len(latencies)
		if current >= pctls[j] {
			data[j] = latencies[i]
			j++
		}
	}
	fmt.Printf("\nLatency distribution:\n")
	for i := 0; i < len(pctls); i++ {
		if data[i] > 0 {
			fmt.Printf("  %v%% in %4.4f secs.\n", pctls[i], data[i])
		}
	}
}

func (b *Boom) printHistogram() {
	bc := 10
	buckets := make([]float64, bc+1)
	counts := make([]int, bc+1)
	fastest := latencies[0]
	slowest := latencies[len(latencies)-1]
	// TODO: Intead of slowest, 95th percentile etc.
	bs := (slowest - fastest) / float64(bc)
	for i := 0; i < bc; i++ {
		buckets[i] = fastest + bs * float64(i)
	}
	buckets[bc] = slowest
	var bi int
	var max int
	for i := 0; i < len(latencies); {
		if latencies[i] <= buckets[bi] {
			i++
			counts[bi]++
			if max < counts[bi] {
				max = counts[bi]
			}
		} else if bi < len(buckets) - 1 {
			bi++
		}
	}
	fmt.Printf("\nResponse time histogram:\n")
	for i := 0; i < len(buckets); i++ {
		// Normalize bar lengths.
		var barLen int
		if max > 0 {
			barLen = counts[i] * 40 / max
		}
		var bar string = ""
		if barLen > 0 {
			bar = strings.Repeat("#", barLen)
		}
		fmt.Printf("  %4.3f [%v]\t|%v\n", buckets[i], counts[i], bar)
	}

}

// Prints status code distribution.
func (b *Boom) printStatusCodes() {
	fmt.Printf("\nStatus code distribution:\n")
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
