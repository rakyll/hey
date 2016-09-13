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

package requester

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	barChar = "∎"
)

type report struct {
	avgTotal float64
	average  float64
	rps      float64

	trace    bool //if trace is set, the following fields will be filled
	avgConn  float64
	avgDns   float64
	avgReq   float64
	avgRes   float64
	avgDelay float64
	conn     []float64
	dns      []float64
	req      []float64
	res      []float64
	delay    []float64

	results chan *result
	total   time.Duration

	errorDist      map[string]int
	statusCodeDist map[int]int
	lats           []float64
	sizeTotal      int64

	output string
}

func newReport(size int, results chan *result, output string, total time.Duration, trace bool) *report {
	return &report{
		output:         output,
		results:        results,
		total:          total,
		trace:          trace,
		statusCodeDist: make(map[int]int),
		errorDist:      make(map[string]int),
	}
}

func (r *report) finalize() {
	for {
		select {
		case res := <-r.results:
			if res.err != nil {
				r.errorDist[res.err.Error()]++
			} else {
				r.lats = append(r.lats, res.duration.Seconds())
				r.avgTotal += res.duration.Seconds()

				if r.trace {
					r.conn = append(r.conn, res.connDuration.Seconds())
					r.avgConn += res.connDuration.Seconds()

					r.delay = append(r.delay, res.delayDuration.Seconds())
					r.avgDelay += res.delayDuration.Seconds()

					r.dns = append(r.dns, res.dnsDuration.Seconds())
					r.avgDns += res.dnsDuration.Seconds()

					r.req = append(r.req, res.reqDuration.Seconds())
					r.avgReq += res.reqDuration.Seconds()

					r.res = append(r.res, res.resDuration.Seconds())
					r.avgRes += res.resDuration.Seconds()
				}

				r.statusCodeDist[res.statusCode]++
				if res.contentLength > 0 {
					r.sizeTotal += res.contentLength
				}
			}
		default:
			r.rps = float64(len(r.lats)) / r.total.Seconds()
			r.average = r.avgTotal / float64(len(r.lats))
			if r.trace {
				r.avgConn = r.avgConn / float64(len(r.lats))
				r.avgDelay = r.avgDelay / float64(len(r.lats))
				r.avgDns = r.avgDns / float64(len(r.lats))
				r.avgReq = r.avgReq / float64(len(r.lats))
				r.avgRes = r.avgRes / float64(len(r.lats))
			}
			r.print()
			return
		}
	}
}

func (r *report) print() {

	if r.output == "csv" {
		r.printCSV()
		return
	}

	if len(r.lats) > 0 {
		var slowest, fastest float64
		sort.Float64s(r.lats)
		fastest = r.lats[0]
		slowest = r.lats[len(r.lats)-1]
		fmt.Printf("Summary:\n")
		fmt.Printf("  Total:\t%4.4f secs\n", r.total.Seconds())
		fmt.Printf("  Slowest:\t%4.4f secs\n", slowest)
		fmt.Printf("  Fastest:\t%4.4f secs\n", fastest)
		fmt.Printf("  Average:\t%4.4f secs\n", r.average)
		fmt.Printf("  Requests/sec:\t%4.4f\n", r.rps)
		if r.sizeTotal > 0 {
			fmt.Printf("  Total data:\t%d bytes\n", r.sizeTotal)
			fmt.Printf("  Size/request:\t%d bytes\n", r.sizeTotal/int64(len(r.lats)))
		}
		r.printStatusCodes()
		printHistogram(r.lats, fastest, slowest)
		r.printLatencies()

		if len(r.errorDist) > 0 {
			r.printErrors()
		}

		if r.trace {
			sort.Float64s(r.dns)
			sort.Float64s(r.conn)
			sort.Float64s(r.delay)
			sort.Float64s(r.res)
			sort.Float64s(r.req)

			fmt.Printf("\n\nDetailed Report:\n")

			slowest, fastest = r.conn[len(r.conn)-1], r.conn[0]
			printSection("DNS+dialup", r.avgConn, fastest, slowest)
			printHistogram(r.conn, fastest, slowest)

			if r.avgDns > 0 {
				slowest, fastest = r.dns[len(r.dns)-1], r.dns[0]
				printSection("DNS", r.avgDns, fastest, slowest)
				printHistogram(r.dns, fastest, slowest)
			}

			slowest, fastest = r.req[len(r.req)-1], r.req[0]
			printSection("Request Write", r.avgReq, fastest, slowest)
			printHistogram(r.req, fastest, slowest)

			slowest, fastest = r.delay[len(r.delay)-1], r.delay[0]
			printSection("Respone Wait", r.avgDelay, fastest, slowest)
			printHistogram(r.delay, fastest, slowest)

			slowest, fastest = r.res[len(r.res)-1], r.res[0]
			printSection("Respone Read", r.avgRes, fastest, slowest)
			printHistogram(r.res, fastest, slowest)
		}
	}
}

func printSection(tag string, avg, fastest, slowest float64) {
	fmt.Printf("\n%s\n", tag)
	fmt.Printf("  Average:\t%4.4f secs\n", avg)
	fmt.Printf("  Fastest:\t%4.4f secs\n", fastest)
	fmt.Printf("  Slowest:\t%4.4f secs\n", slowest)
}

func (r *report) printCSV() {
	fmt.Printf("response-time")
	if r.trace {
		fmt.Printf(",DNS+Dialup,DNS,request-write,respone-wait,respone-read")
	}
	fmt.Println()
	for i, val := range r.lats {
		fmt.Printf("%4.4f", val)
		if r.trace {
			fmt.Printf(",%4.4f,%4.4f,%4.4f,%4.4f,%4.4f", r.conn[i], r.dns[i], r.req[i],
				r.delay[i], r.res[i])
		}
		fmt.Println()
	}
}

// Prints percentile latencies.
func (r *report) printLatencies() {
	pctls := []int{10, 25, 50, 75, 90, 95, 99}
	data := make([]float64, len(pctls))
	j := 0
	for i := 0; i < len(r.lats) && j < len(pctls); i++ {
		current := i * 100 / len(r.lats)
		if current >= pctls[j] {
			data[j] = r.lats[i]
			j++
		}
	}
	fmt.Printf("\nLatency distribution:\n")
	for i := 0; i < len(pctls); i++ {
		if data[i] > 0 {
			fmt.Printf("  %v%% in %4.4f secs\n", pctls[i], data[i])
		}
	}
}

func printHistogram(lats []float64, fastest, slowest float64) {
	bc := 10
	buckets := make([]float64, bc+1)
	counts := make([]int, bc+1)
	bs := (slowest - fastest) / float64(bc)
	for i := 0; i < bc; i++ {
		buckets[i] = fastest + bs*float64(i)
	}
	buckets[bc] = slowest
	var bi int
	var max int
	for i := 0; i < len(lats); {
		if lats[i] <= buckets[bi] {
			i++
			counts[bi]++
			if max < counts[bi] {
				max = counts[bi]
			}
		} else if bi < len(buckets)-1 {
			bi++
		}
	}
	fmt.Printf("\nHistogram:\n")
	for i := 0; i < len(buckets); i++ {
		// Normalize bar lengths.
		var barLen int
		if max > 0 {
			barLen = counts[i] * 40 / max
		}
		fmt.Printf("  %4.3f [%v]\t|%v\n", buckets[i], counts[i], strings.Repeat(barChar, barLen))
	}
}

// Prints status code distribution.
func (r *report) printStatusCodes() {
	fmt.Printf("\nStatus code distribution:\n")
	for code, num := range r.statusCodeDist {
		fmt.Printf("  [%d]\t%d responses\n", code, num)
	}
}

func (r *report) printErrors() {
	fmt.Printf("\nError distribution:\n")
	for err, num := range r.errorDist {
		fmt.Printf("  [%d]\t%s\n", num, err)
	}
}
