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
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestStopWhileWaitingForRateLimit(t *testing.T) {
	var count int64
	first := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt64(&count, 1) == 1 {
			first <- struct{}{}
		}
	}))
	defer server.Close()
	req, _ := http.NewRequest("GET", server.URL, nil)
	work := &Work{Request: req, N: 10, C: 1, QPS: 2, Writer: io.Discard}
	work.Init()
	done := make(chan struct{})
	go func() { work.Run(); close(done) }()
	select {
	case <-first:
	case <-time.After(5 * time.Second):
		t.Fatal("first request never arrived")
	}
	// Let the first response complete and the worker wait for its next tick.
	time.Sleep(50 * time.Millisecond)
	work.Stop()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("rate-limited worker did not stop")
	}
	if got := atomic.LoadInt64(&count); got != 1 {
		t.Fatalf("sent %d requests; want only the request made before Stop", got)
	}
}
