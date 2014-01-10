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
)

func (b *Boom) Print() {
	total := b.end.Sub(b.start)
	var fastest, slowest float64
	rps := float64(b.N) / total.Seconds()
	lat := b.report.latencies
	if len(lat) > 0 {
		fastest = lat[0]
		slowest = lat[len(lat)-1]
		sort.Float64s(b.report.latencies)
		fmt.Printf("\nSummary:\n")
		fmt.Printf("  Total:\t%4.4f secs.\n", total.Seconds())
		fmt.Printf("  Slowest:\t%4.4f secs.\n", slowest)
		fmt.Printf("  Fastest:\t%4.4f secs.\n", fastest)
		fmt.Printf("  Average:\t%4.4f secs.\n", b.report.avgTotal/float64(b.N))
		fmt.Printf("  Requests/sec:\t%4.4f\n", rps)
		fmt.Printf("  Speed index:\t%v\n", speedIndex(rps))
		b.printStatusCodes()
		b.printHistogram(&lat)
		b.printLatencies(&lat)
	}
}

// Prints percentile latencies.
func (b *Boom) printLatencies(latPtr *[]float64) {
	lat := *latPtr
	pctls := []int{10, 25, 50, 75, 90, 95, 99}
	// Sort the array
	data := make([]float64, len(pctls))
	j := 0
	for i := 0; i < len(lat) && j < len(pctls); i++ {
		current := (i + 1) * 100 / len(lat)
		if current >= pctls[j] {
			data[j] = lat[i]
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

func (b *Boom) printHistogram(latPtr *[]float64) {
	lat := *latPtr
	bc := 10
	buckets := make([]float64, bc+1)
	counts := make([]int, bc+1)
	fastest := lat[0]
	slowest := lat[len(lat)-1]
	bs := (slowest - fastest) / float64(bc)
	for i := 0; i < bc; i++ {
		buckets[i] = fastest + bs*float64(i)
	}
	buckets[bc] = slowest
	var bi int
	var max int
	for i := 0; i < len(lat); {
		if lat[i] <= buckets[bi] {
			i++
			counts[bi]++
			if max < counts[bi] {
				max = counts[bi]
			}
		} else if bi < len(buckets)-1 {
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
		fmt.Printf("  %4.3f [%v]\t|%v\n", buckets[i], counts[i], strings.Repeat("#", barLen))
	}
}

// Prints status code distribution.
func (b *Boom) printStatusCodes() {
	fmt.Printf("\nStatus code distribution:\n")
	for code, num := range b.report.statusCodeDist {
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
