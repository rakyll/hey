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
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"
)

func (b *Boom) Run() {
	b.results = make(chan *result, b.N)
	if b.Output == "" {
		b.bar = newPb(b.N)
	}
	b.rpt = newReport(b.N, b.results, b.Output)
	b.run()
}

func (b *Boom) worker(ch chan *http.Request) {
	host, _, _ := net.SplitHostPort(b.Req.OriginalHost)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: b.AllowInsecure, ServerName: host},
	}
	client := &http.Client{Transport: tr}
	for req := range ch {
		s := time.Now()
		resp, err := client.Do(req)
		code := 0
		if resp != nil {
			code = resp.StatusCode
			// cleanup body, so the socket can be reusable
			resp.Body.Close()
		}
		if b.bar != nil {
			b.bar.Increment()
		}
		b.results <- &result{
			statusCode: code,
			duration:   time.Now().Sub(s),
			err:        err,
		}
	}
}

func (b *Boom) run() {
	var wg sync.WaitGroup
	wg.Add(b.C)

	var throttle <-chan time.Time
	if b.Qps > 0 {
		throttle = time.Tick(time.Duration(1e6/(b.Qps)) * time.Microsecond)
	}

	start := time.Now()
	jobs := make(chan *http.Request, b.N)
	// Start workers.
	for i := 0; i < b.C; i++ {
		go func() {
			b.worker(jobs)
			wg.Done()
		}()
	}

	// Start sending jobs to the workers.
	for i := 0; i < b.N; i++ {
		if b.Qps > 0 {
			<-throttle
		}
		jobs <- b.Req.Request()
	}
	close(jobs)

	wg.Wait()
	if b.bar != nil {
		b.bar.Finish()
	}
	b.rpt.finalize(time.Now().Sub(start))
}
