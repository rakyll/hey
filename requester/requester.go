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
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

const heyUA = "hey/0.0.1"

// Max size of the buffer of result channel.
const maxResult = 1000000
const maxIdleConn = 500

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

	RequestBody []byte

	// N is the total number of requests to make.
	N int

	// C is the concurrency level, the number of concurrent workers to run.
	C int

	// H2 is an option to make HTTP/2 requests
	H2 bool

	// Timeout in seconds.
	Timeout int

	// Qps is the request rate limit per worker.
	QPS int

	// DisableCompression is an option to disable compression in response
	DisableCompression bool

	// DisableKeepAlives is an option to prevents re-use of TCP connections between different HTTP requests
	DisableKeepAlives bool

	// DisableRedirects is an option to prevent the following of HTTP redirects
	DisableRedirects bool

	// Output represents the output type. If "csv" is provided, the
	// output will be dumped as a csv stream.
	Output string

	// ProxyAddr is the address of HTTP proxy server in the format on "host:port".
	// Optional.
	ProxyAddr *url.URL

	// Writer is where results will be written. If nil, results are written to stdout.
	Writer io.Writer

	results chan *result
	stopCh  chan struct{}
	start   time.Time

	report *report
}

func (b *Work) writer() io.Writer {
	if b.Writer == nil {
		return os.Stdout
	}
	return b.Writer
}

// Run makes all the requests, prints the summary. It blocks until
// all work is done.
func (b *Work) Run() {
	// append hey's user agent
	ua := b.Request.UserAgent()
	if ua == "" {
		ua = heyUA
	} else {
		ua += " " + heyUA
	}

	b.results = make(chan *result, min(b.C*1000, maxResult))
	b.stopCh = make(chan struct{}, b.C)
	b.start = time.Now()
	b.report = newReport(b.writer(), b.results, b.Output, b.N)
	// Run the reporter first, it polls the result channel until it is closed.
	go func() {
		runReporter(b.report)
	}()
	b.runWorkers()
	b.Finish()
}

func (b *Work) Stop() {
	// Send stop signal so that workers can stop gracefully.
	for i := 0; i < b.C; i++ {
		b.stopCh <- struct{}{}
	}
}

func (b *Work) Finish() {
	close(b.results)
	total := time.Now().Sub(b.start)
	// Wait until the reporter is done.
	<-b.report.done
	b.report.finalize(total)
}

func (b *Work) makeRequest(c *http.Client) {
	s := time.Now()
	var size int64
	var code int
	var dnsStart, connStart, resStart, reqStart, delayStart time.Time
	var dnsDuration, connDuration, resDuration, reqDuration, delayDuration time.Duration
	req := cloneRequest(b.Request, b.RequestBody)
	trace := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			dnsDuration = time.Now().Sub(dnsStart)
		},
		GetConn: func(h string) {
			connStart = time.Now()
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			connDuration = time.Now().Sub(connStart)
			reqStart = time.Now()
		},
		WroteRequest: func(w httptrace.WroteRequestInfo) {
			reqDuration = time.Now().Sub(reqStart)
			delayStart = time.Now()
		},
		GotFirstResponseByte: func() {
			delayDuration = time.Now().Sub(delayStart)
			resStart = time.Now()
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	resp, err := c.Do(req)
	if err == nil {
		size = resp.ContentLength
		code = resp.StatusCode
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}
	t := time.Now()
	resDuration = t.Sub(resStart)
	finish := t.Sub(s)
	b.results <- &result{
		statusCode:    code,
		duration:      finish,
		err:           err,
		contentLength: size,
		connDuration:  connDuration,
		dnsDuration:   dnsDuration,
		reqDuration:   reqDuration,
		resDuration:   resDuration,
		delayDuration: delayDuration,
	}
}

func (b *Work) runWorker(client *http.Client, n int) {
	var throttle <-chan time.Time
	var wg sync.WaitGroup
	if b.QPS > 0 {
		throttle = time.Tick(time.Duration(1e6/(b.QPS)) * time.Microsecond)
		wg.Add(n)
	}

	if b.DisableRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	for i := 0; i < n; i++ {
		// Check if application is stopped. Do not send into a closed channel.
		select {
		case <-b.stopCh:
			return
		default:
			if b.QPS > 0 {
				go func() {
					b.makeRequest(client)
					wg.Done()
				}()
				<-throttle
			} else {
				b.makeRequest(client)
			}
		}
	}
	if b.QPS > 0 {
		wg.Wait()
	}
}

func (b *Work) runWorkers() {
	var wg sync.WaitGroup
	wg.Add(b.C)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		MaxIdleConnsPerHost: min(b.C, maxIdleConn),
		DisableCompression:  b.DisableCompression,
		DisableKeepAlives:   b.DisableKeepAlives,
		Proxy:               http.ProxyURL(b.ProxyAddr),
	}
	if b.H2 {
		http2.ConfigureTransport(tr)
	} else {
		tr.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
	}
	client := &http.Client{Transport: tr, Timeout: time.Duration(b.Timeout) * time.Second}

	// Ignore the case where b.N % b.C != 0.
	for i := 0; i < b.C; i++ {
		go func() {
			b.runWorker(client, b.N/b.C)
			wg.Done()
		}()
	}
	wg.Wait()
}

// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request, body []byte) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	if len(body) > 0 {
		r2.Body = ioutil.NopCloser(bytes.NewReader(body))
	}
	return r2
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
