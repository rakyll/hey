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
	"encoding/csv"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
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
		t.Errorf("Expected to boom 20 times, found %v", count)
	}
}

func TestCSV(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	outputBuffer := &bytes.Buffer{}

	req, _ := http.NewRequest("GET", server.URL, nil)
	w := &Work{
		Request: req,
		N:       10,
		C:       2,
		Writer:  outputBuffer,
		Output:  "csv",
	}
	w.Run()

	csvReader := csv.NewReader(outputBuffer)
	records, err := csvReader.ReadAll()
	if err != nil {
		t.Errorf("parsing csv: %s", err)
	}

	nExpectedRecords := w.N + 1 // for header
	if got, want := len(records), nExpectedRecords; got != want {
		t.Errorf("num records = %d; want %d", len(records), nExpectedRecords)
	}

	expectedHeadersLine := "response-time,DNS+dialup,DNS,Request-write,Response-delay,Response-read,start-time"
	if got, want := strings.Join(records[0], ","), expectedHeadersLine; got != want {
		t.Errorf("headers = %s; want %s", records[0], expectedHeadersLine)
	}

	// last column is start-time, can be parsed as a RFC339Nano
	startTimeColumn := len(records[0]) - 1
	startTime1 := records[1][startTimeColumn]
	_, err = time.Parse(time.RFC3339Nano, startTime1)
	if err != nil {
		t.Errorf("parsing start time as integer: %s", err)
	}

	const nExpectedColumns = 7
	for i, row := range records {
		if got, want := len(row), nExpectedColumns; got != want {
			t.Errorf("row %d column count = %d; want %d", i, len(row), nExpectedColumns)
		}
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
		if count > 1 {
			t.Errorf("Expected to work 1 times, found %v", count)
		}
		wg.Done()
	})
	go w.Run()
	wg.Wait()
}

func TestRequest(t *testing.T) {
	var uri, contentType, some, method, auth string
	handler := func(w http.ResponseWriter, r *http.Request) {
		uri = r.RequestURI
		method = r.Method
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
		body, _ := ioutil.ReadAll(r.Body)
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
