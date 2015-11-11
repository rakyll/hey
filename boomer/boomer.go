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

// Package boomer provides commands to run load tests and display results.
package boomer

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rakyll/pb"
)

// client is the http.Client that will be used to make all requests
// to the destination.
var client *http.Client

type result struct {
	err           error
	statusCode    int
	duration      time.Duration
	contentLength int64
}

type ReqOpts struct {
	Method   string
	URL      string
	Header   http.Header
	Body     string
	Username string
	Password string
	ReadAll  bool
}

// Creates a req object from req options
func (r *ReqOpts) Request() *http.Request {
	req, _ := http.NewRequest(r.Method, r.URL, strings.NewReader(r.Body))
	req.Header = r.Header
	if r.Username != "" && r.Password != "" {
		req.SetBasicAuth(r.Username, r.Password)
	}
	return req
}

type Boomer struct {
	// Req represents the options of the request to be made.
	// TODO(jbd): Make it work with an http.Request instead.
	Req *ReqOpts

	// N is the total number of requests to make.
	N int

	// C is the concurrency level, the number of concurrent workers to run.
	C int

	// Timeout in seconds.
	Timeout int

	// Qps is the rate limit.
	Qps int

	// AllowInsecure is an option to allow insecure TLS/SSL certificates.
	AllowInsecure bool

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

	bar     *pb.ProgressBar
	results chan *result
}

func (b *Boomer) startProgress() {
	if b.Output != "" {
		return
	}
	b.bar = pb.New(b.N)
	b.bar.Format("Bom !")
	b.bar.Start()
}

func (b *Boomer) finalizeProgress() {
	if b.Output != "" {
		return
	}
	b.bar.Finish()
}

func (b *Boomer) incProgress() {
	if b.Output != "" {
		return
	}
	b.bar.Increment()
}

// Run makes all the requests, prints the summary. It blocks until
// all work is done.
func (b *Boomer) Run() {
	b.results = make(chan *result, b.N)
	b.startProgress()

	start := time.Now()
	b.runWorkers()
	b.finalizeProgress()

	newReport(b.N, b.results, b.Output, time.Now().Sub(start)).finalize()
	close(b.results)
}

func (b *Boomer) runWorker(wg *sync.WaitGroup, ch chan *http.Request, readAll bool) {
	for req := range ch {
		s := time.Now()

		var code int
		var size int64

		resp, err := client.Do(req)
		if err == nil {
			size = resp.ContentLength
			code = resp.StatusCode
			if readAll {
				_, err = io.Copy(ioutil.Discard, resp.Body)
			}
			resp.Body.Close()
		}

		wg.Done()
		b.incProgress()
		b.results <- &result{
			statusCode:    code,
			duration:      time.Now().Sub(s),
			err:           err,
			contentLength: size,
		}
	}
}

func (b *Boomer) runWorkers() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: b.AllowInsecure,
		},
		DisableCompression: b.DisableCompression,
		DisableKeepAlives:  b.DisableKeepAlives,
		// TODO(jbd): Add dial timeout.
		TLSHandshakeTimeout: time.Duration(b.Timeout) * time.Millisecond,
		Proxy:               http.ProxyURL(b.ProxyAddr),
	}
	client = &http.Client{Transport: tr}

	var wg sync.WaitGroup
	wg.Add(b.N)

	var throttle <-chan time.Time
	if b.Qps > 0 {
		throttle = time.Tick(time.Duration(1e6/(b.Qps)) * time.Microsecond)
	}

	jobsch := make(chan *http.Request, b.N)
	for i := 0; i < b.C; i++ {
		go b.runWorker(&wg, jobsch, b.Req.ReadAll)
	}

	for i := 0; i < b.N; i++ {
		if b.Qps > 0 {
			<-throttle
		}
		jobsch <- b.Req.Request()
	}
	close(jobsch)

	wg.Wait()
}
