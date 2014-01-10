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
	"io"
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
	b.results = make(chan *result, b.N)
	b.bar = pb.StartNew(b.N)
	b.start = time.Now()
}

func (b *Boom) teardown() {
	b.end = time.Now()
	b.bar.Finish()
}

func (b *Boom) run() {
	var wg sync.WaitGroup
	ch := make(chan bool, b.C)
	// Create c amount worker goroutines.
	for i := 0; i < b.C; i++ {
		wg.Add(1)
		go b.worker(ch, &wg)
	}
	// Push requests to channel.
	for i := 0; i < b.N; i++ {
		ch <- true
	}
	close(ch)
	// Wait all goroutines to finish.
	wg.Wait()
}

func (b *Boom) worker(ch chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	buffer := make([]byte, 1024*1024*2)

	// Pick one request from channel and process.
	for _ = range ch {
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

		if resp != nil {
			io.ReadFull(resp.Body, buffer)
			resp.Body.Close()
		}

		b.bar.Increment()
	}
}
