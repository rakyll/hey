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
	"io"
	"sort"
	"strings"
	"time"
)

const (
	barChar = "∎"
)

type report struct {
	avgTotal float64
	fastest  float64
	slowest  float64
	average  float64
	rps      float64

	avgConn   float64
	avgDNS    float64
	avgReq    float64
	avgRes    float64
	avgDelay  float64
	connLats  []float64
	dnsLats   []float64
	reqLats   []float64
	resLats   []float64
	delayLats []float64

	results chan *result
	total   time.Duration

	errorDist      map[string]int
	statusCodeDist map[int]int
	lats           []float64
	startTimes     []time.Time
	sizeTotal      int64

	output string

	w io.Writer
}

func newReport(w io.Writer, size int, results chan *result, output string, total time.Duration) *report {
	return &report{
		output:         output,
		results:        results,
		total:          total,
		statusCodeDist: make(map[int]int),
		errorDist:      make(map[string]int),
		w:              w,
	}
}

func (r *report) finalize() {
	for res := range r.results {
		if res.err != nil {
			r.errorDist[res.err.Error()]++
		} else {
			r.startTimes = append(r.startTimes, res.startTime)
			r.lats = append(r.lats, res.duration.Seconds())
			r.avgTotal += res.duration.Seconds()
			r.avgConn += res.connDuration.Seconds()
			r.avgDelay += res.delayDuration.Seconds()
			r.avgDNS += res.dnsDuration.Seconds()
			r.avgReq += res.reqDuration.Seconds()
			r.avgRes += res.resDuration.Seconds()
			r.connLats = append(r.connLats, res.connDuration.Seconds())
			r.dnsLats = append(r.dnsLats, res.dnsDuration.Seconds())
			r.reqLats = append(r.reqLats, res.reqDuration.Seconds())
			r.delayLats = append(r.delayLats, res.delayDuration.Seconds())
			r.resLats = append(r.resLats, res.resDuration.Seconds())
			r.statusCodeDist[res.statusCode]++
			if res.contentLength > 0 {
				r.sizeTotal += res.contentLength
			}
		}
	}
	r.rps = float64(len(r.lats)) / r.total.Seconds()
	r.average = r.avgTotal / float64(len(r.lats))
	r.avgConn = r.avgConn / float64(len(r.lats))
	r.avgDelay = r.avgDelay / float64(len(r.lats))
	r.avgDNS = r.avgDNS / float64(len(r.lats))
	r.avgReq = r.avgReq / float64(len(r.lats))
	r.avgRes = r.avgRes / float64(len(r.lats))
	r.print()
}

func (r *report) printCSV() {
	r.printf("response-time,DNS+dialup,DNS,Request-write,Response-delay,Response-read,start-time\n")
	for i, val := range r.lats {
		r.printf("%4.4f,%4.4f,%4.4f,%4.4f,%4.4f,%4.4f,%s\n",
			val, r.connLats[i], r.dnsLats[i], r.reqLats[i], r.delayLats[i], r.resLats[i], r.startTimes[i].UTC().Format(time.RFC3339Nano))
	}
}

func (r *report) print() {
	if r.output == "csv" {
		r.printCSV()
		return
	}

	if len(r.lats) > 0 {
		sort.Float64s(r.lats)
		r.fastest = r.lats[0]
		r.slowest = r.lats[len(r.lats)-1]
		r.printf("Summary:\n")
		r.printf("  Total:\t%4.4f secs\n", r.total.Seconds())
		r.printf("  Slowest:\t%4.4f secs\n", r.slowest)
		r.printf("  Fastest:\t%4.4f secs\n", r.fastest)
		r.printf("  Average:\t%4.4f secs\n", r.average)
		r.printf("  Requests/sec:\t%4.4f\n", r.rps)
		if r.sizeTotal > 0 {
			r.printf("  Total data:\t%d bytes\n", r.sizeTotal)
			r.printf("  Size/request:\t%d bytes\n", r.sizeTotal/int64(len(r.lats)))
		}
		r.printHistogram()
		r.printLatencies()
		r.printf("\nDetails (average, fastest, slowest):")
		r.printSection("DNS+dialup", r.avgConn, r.connLats)
		r.printSection("DNS-lookup", r.avgDNS, r.dnsLats)
		r.printSection("req write", r.avgReq, r.reqLats)
		r.printSection("resp wait", r.avgDelay, r.delayLats)
		r.printSection("resp read", r.avgRes, r.resLats)
		r.printStatusCodes()
	}
	if len(r.errorDist) > 0 {
		r.printErrors()
	}
	r.printf("\n")
}

// printSection prints details for http-trace fields
func (r *report) printSection(tag string, avg float64, lats []float64) {
	sort.Float64s(lats)
	fastest, slowest := lats[0], lats[len(lats)-1]
	r.printf("\n  %s:\t", tag)
	r.printf(" %4.4f secs, %4.4f secs, %4.4f secs", avg, fastest, slowest)
}

// printLatencies prints percentile latencies.
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
	r.printf("\nLatency distribution:\n")
	for i := 0; i < len(pctls); i++ {
		if data[i] > 0 {
			r.printf("  %v%% in %4.4f secs\n", pctls[i], data[i])
		}
	}
}

func (r *report) printHistogram() {
	bc := 10
	buckets := make([]float64, bc+1)
	counts := make([]int, bc+1)
	bs := (r.slowest - r.fastest) / float64(bc)
	for i := 0; i < bc; i++ {
		buckets[i] = r.fastest + bs*float64(i)
	}
	buckets[bc] = r.slowest
	var bi int
	var max int
	for i := 0; i < len(r.lats); {
		if r.lats[i] <= buckets[bi] {
			i++
			counts[bi]++
			if max < counts[bi] {
				max = counts[bi]
			}
		} else if bi < len(buckets)-1 {
			bi++
		}
	}
	r.printf("\nResponse time histogram:\n")
	for i := 0; i < len(buckets); i++ {
		// Normalize bar lengths.
		var barLen int
		if max > 0 {
			barLen = (counts[i]*40 + max/2) / max
		}
		r.printf("  %4.3f [%v]\t|%v\n", buckets[i], counts[i], strings.Repeat(barChar, barLen))
	}
}

// printStatusCodes prints status code distribution.
func (r *report) printStatusCodes() {
	r.printf("\n\nStatus code distribution:\n")
	for code, num := range r.statusCodeDist {
		r.printf("  [%d]\t%d responses\n", code, num)
	}
}

func (r *report) printErrors() {
	r.printf("\nError distribution:\n")
	for err, num := range r.errorDist {
		r.printf("  [%d]\t%s\n", num, err)
	}
}

func (r *report) printf(s string, v ...interface{}) {
	fmt.Fprintf(r.w, s, v...)
}
