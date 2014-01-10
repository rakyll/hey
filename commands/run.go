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
	"net/http"
	"sync"
	"time"

	"github.com/cheggaaa/pb"
)

func (b *Boom) Run() {
	b.Req.Header.Add("cache-control", "no-cache")
	b.init()
	b.run()
	b.teardown()
	b.Print()
}

func (b *Boom) init() {
	if b.Client == nil {
		b.Client = &http.Client{}
	}
	b.results = make(chan *result, b.C)
	b.bar = pb.StartNew(b.N)
	b.report.statusCodeDist = make(map[int]int)
	b.start = time.Now()
	b.timedOut = false
}

func (b *Boom) teardown() {
	b.end = time.Now()
	b.bar.Finish()
}

func (b *Boom) worker(jobs chan bool, wg *sync.WaitGroup, timeout <-chan time.Time) {
	defer wg.Done()
WORKER_LOOP:
	for !b.timedOut {
		select {
		case _, chOpen := <-jobs:
			if !chOpen {
				break WORKER_LOOP
			}
			s := time.Now()
			resp, err := b.Client.Do(b.Req)
			code := 0
			if resp != nil {
				code = resp.StatusCode
			}
			b.results <- &result{
				statusCode: code,
				duration:   time.Now().Sub(s),
				err:        err,
			}
			b.bar.Increment()
		case <-timeout:
			b.timedOut = true
		}
	}
}

func (b *Boom) collector() {
	for {
		select {
		case r := <-b.results:
			b.report.latencies = append(b.report.latencies, r.duration.Seconds())
			b.report.statusCodeDist[r.statusCode]++
			b.report.avgTotal += r.duration.Seconds()
		}
	}
}

func (b *Boom) run() {
	jobs := make(chan bool, b.N)
	var wg sync.WaitGroup
	// Start collector.
	go b.collector()
	// Start throttler if rate limit is specified.
	var throttle <-chan time.Time
	if b.Q > 0 {
		throttle = time.Tick(time.Duration(1e6/b.Q) * time.Microsecond)
	}
	// Start timeout counter if time limit is specified.
	var timeout <-chan time.Time
	if b.S > 0 {
		timeout = time.After(time.Duration(b.S) * time.Second)
	}
	// Start workers.
	for i := 0; i < b.C; i++ {
		wg.Add(1)
		go b.worker(jobs, &wg, timeout)
	}
	// Start sending requests.
	for i := 0; i < b.N; i++ {
		if b.Q > 0 {
			<-throttle
		}
		jobs <- true
	}
	close(jobs)
	wg.Wait()
}
