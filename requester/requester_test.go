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
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestN(t *testing.T) {
	var count int64
	handler := func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, int64(1))
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	w := &Work{
		Request: req,
		N:       20,
		C:       2,
	}
	w.Run()
	if count != 20 {
		t.Errorf("Expected to send 20 requests, found %v", count)
	}
}

func TestQps(t *testing.T) {
	var wg sync.WaitGroup
	var count int64
	handler := func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, int64(1))
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	w := &Work{
		Request: req,
		N:       20,
		C:       2,
		QPS:     1,
	}
	wg.Add(1)
	time.AfterFunc(time.Second, func() {
		if count > 2 {
			t.Errorf("Expected to work at most 2 times, found %v", count)
		}
		wg.Done()
	})
	go w.Run()
	wg.Wait()
}

func TestRequest(t *testing.T) {
	var uri, contentType, some, auth string
	handler := func(w http.ResponseWriter, r *http.Request) {
		uri = r.RequestURI
		contentType = r.Header.Get("Content-type")
		some = r.Header.Get("X-some")
		auth = r.Header.Get("Authorization")
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	header := make(http.Header)
	header.Add("Content-type", "text/html")
	header.Add("X-some", "value")
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header = header
	req.SetBasicAuth("username", "password")
	w := &Work{
		Request: req,
		N:       1,
		C:       1,
	}
	w.Run()
	if uri != "/" {
		t.Errorf("Uri is expected to be /, %v is found", uri)
	}
	if contentType != "text/html" {
		t.Errorf("Content type is expected to be text/html, %v is found", contentType)
	}
	if some != "value" {
		t.Errorf("X-some header is expected to be value, %v is found", some)
	}
	if auth != "Basic dXNlcm5hbWU6cGFzc3dvcmQ=" {
		t.Errorf("Basic authorization is not properly set")
	}
}

func TestBody(t *testing.T) {
	var count int64
	handler := func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) == "Body" {
			atomic.AddInt64(&count, 1)
		}
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	req, _ := http.NewRequest("POST", server.URL, bytes.NewBuffer([]byte("Body")))
	w := &Work{
		Request:     req,
		RequestBody: []byte("Body"),
		N:           10,
		C:           1,
	}
	w.Run()
	if count != 10 {
		t.Errorf("Expected to work 10 times, found %v", count)
	}
}

// Status counts and averages include successful results beyond the sample cap.
func TestReportBeyondSampleLimit(t *testing.T) {
	results := make(chan *result, 3)
	for i, status := range []int{http.StatusOK, http.StatusInternalServerError} {
		d := time.Duration(2*i+1) * time.Second
		results <- &result{statusCode: status, duration: d, connDuration: 2 * d,
			dnsDuration: 3 * d, reqDuration: 4 * d, delayDuration: 5 * d,
			resDuration: 6 * d, contentLength: int64(10 * (2*i + 1))}
	}
	results <- &result{err: errors.New("request failed")}
	close(results)
	r := newReport(io.Discard, results, "{{.NumRes}}", 3)
	runReporter(r, 1)
	r.finalize(time.Second)
	got := r.snapshot()
	if got.NumRes != 3 || got.Rps != 3 || got.ErrorDist["request failed"] != 1 {
		t.Errorf("request/error accounting = %d/%g/%v, want 3/3/one failure", got.NumRes, got.Rps, got.ErrorDist)
	}
	if len(got.StatusCodeDist) != 2 || got.StatusCodeDist[http.StatusOK] != 1 || got.StatusCodeDist[http.StatusInternalServerError] != 1 {
		t.Errorf("status distribution = %v, want one OK and one server error", got.StatusCodeDist)
	}
	for i, average := range []float64{got.Average, got.AvgConn / 2, got.AvgDNS / 3, got.AvgReq / 4, got.AvgDelay / 5, got.AvgRes / 6} {
		if average != 2 {
			t.Errorf("normalized average %d = %g, want 2", i, average)
		}
	}
	if got.SizeReq != 20 || got.SizeTotal != 40 || len(got.Lats) != 1 || len(got.StatusCodes) != 1 {
		t.Errorf("size/sample accounting = %d/%d/%d/%d, want 20/40/1/1", got.SizeReq, got.SizeTotal, len(got.Lats), len(got.StatusCodes))
	}
}

func TestReportNoSuccessfulResults(t *testing.T) {
	for _, requestError := range []error{nil, errors.New("request failed")} {
		var output bytes.Buffer
		results := make(chan *result, 1)
		if requestError != nil {
			results <- &result{err: requestError}
		}
		close(results)
		r := newReport(&output, results, "{{jsonify .StatusCodeDist}}", 1)
		runReporter(r, 1)
		r.finalize(time.Second)
		if output.String() != "null\n" {
			t.Errorf("status JSON for error %v = %q, want null", requestError, output.String())
		}
	}
}
