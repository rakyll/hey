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

// Package requester provides commands to run load tests and display results.
package requester

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

type result struct {
	err           error
	statusCode    int
	duration      time.Duration
	connDuration  time.Duration // connection setup(DNS lookup + Dial up) duration
	dnsDuration   time.Duration // dns lookup duration
	reqDuration   time.Duration // request "write" duration
	resDuration   time.Duration // response "read" duration
	delayDuration time.Duration // delay between response and request
	contentLength int64
}

type Work struct {
	// Request is the request to be made.
	Request *http.Request

	RequestBody string

	// N is the total number of requests to make.
	N int

	// C is the concurrency level, the number of concurrent workers to run.
	C int

	// H2 is an option to make HTTP/2 requests
	H2 bool

	// EnableTrace is an option to enable httpTrace
	EnableTrace bool

	// Timeout in seconds.
	Timeout int

	// Qps is the rate limit.
	Qps int

	// DisableCompression is an option to disable compression in response
	DisableCompression bool

	// DisableKeepAlives is an option to prevents re-use of TCP connections between different HTTP requests
	DisableKeepAlives bool

	// Output represents the output type. If "csv" is provided, the
	// output will be dumped as a csv stream.
	Output string

	// ProxyAddr is the address of HTTP proxy server in the format on "host:port".
	// Optional.
	ProxyAddr *url.URL

	results chan *result
}

// Run makes all the requests, prints the summary. It blocks until
// all work is done.
func (b *Work) Run() {
	b.results = make(chan *result, b.N)

	start := time.Now()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		// TODO(jbd): Progress bar should not be finalized.
		newReport(b.N, b.results, b.Output, time.Now().Sub(start), b.EnableTrace).finalize()
		os.Exit(1)
	}()

	b.runWorkers()
	newReport(b.N, b.results, b.Output, time.Now().Sub(start), b.EnableTrace).finalize()
	close(b.results)
}

func (b *Work) makeRequest(c *http.Client) {
	s := time.Now()
	var size int64
	var code int
	var dnsDuration, connDuration, resDuration, reqDuration, delayTime time.Duration
	req := cloneRequest(b.Request, b.RequestBody)
	if b.EnableTrace {
		trace := &httptrace.ClientTrace{
			DNSStart: func(info httptrace.DNSStartInfo) {
				dnsDuration = time.Now().Sub(s)
			},
			DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
				t := time.Now().Sub(s)
				dnsDuration = time.Duration(t.Nanoseconds() - dnsDuration.Nanoseconds())
			},
			GetConn: func(h string) {
				connDuration = time.Now().Sub(s)
			},
			GotConn: func(connInfo httptrace.GotConnInfo) {
				t := time.Now().Sub(s)
				reqDuration = t
				connDuration = time.Duration(t.Nanoseconds() - connDuration.Nanoseconds())
			},
			WroteRequest: func(w httptrace.WroteRequestInfo) {
				t := time.Now().Sub(s)
				delayTime = t
				reqDuration = time.Duration(t.Nanoseconds() - reqDuration.Nanoseconds())
			},
			GotFirstResponseByte: func() {
				resDuration = time.Now().Sub(s)
				delayTime = time.Duration(resDuration.Nanoseconds() - delayTime.Nanoseconds())
			},
		}
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	}
	resp, err := c.Do(req)
	if err == nil {
		size = resp.ContentLength
		code = resp.StatusCode
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}
	finish := time.Now().Sub(s)
	resDuration = time.Duration(finish.Nanoseconds() - resDuration.Nanoseconds())
	b.results <- &result{
		statusCode:    code,
		duration:      finish,
		err:           err,
		contentLength: size,
		connDuration:  connDuration,
		dnsDuration:   dnsDuration,
		reqDuration:   reqDuration,
		resDuration:   resDuration,
		delayDuration: delayTime,
	}
}

func (b *Work) runWorker(n int) {
	var throttle <-chan time.Time
	if b.Qps > 0 {
		throttle = time.Tick(time.Duration(1e6/(b.Qps)) * time.Microsecond)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		DisableCompression: b.DisableCompression,
		DisableKeepAlives:  b.DisableKeepAlives,
		// TODO(jbd): Add dial timeout.
		TLSHandshakeTimeout: time.Duration(b.Timeout) * time.Millisecond,
		Proxy:               http.ProxyURL(b.ProxyAddr),
	}
	if b.H2 {
		http2.ConfigureTransport(tr)
	} else {
		tr.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
	}
	client := &http.Client{Transport: tr}
	for i := 0; i < n; i++ {
		if b.Qps > 0 {
			<-throttle
		}
		b.makeRequest(client)
	}
}

func (b *Work) runWorkers() {
	var wg sync.WaitGroup
	wg.Add(b.C)

	// Ignore the case where b.N % b.C != 0.
	for i := 0; i < b.C; i++ {
		go func() {
			b.runWorker(b.N / b.C)
			wg.Done()
		}()
	}
	wg.Wait()
}

// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request, body string) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	r2.Body = ioutil.NopCloser(strings.NewReader(body))
	return r2
}
