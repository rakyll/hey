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
	"net/http"
	"strings"
	"time"

	"github.com/rakyll/pb"
)

type result struct {
	err           error
	statusCode    int
	duration      time.Duration
	contentLength int64
}

type ReqOpts struct {
	Method   string
	Url      string
	Header   http.Header
	Body     string
	Username string
	Password string
	// OriginalHost represents the original host name user is provided.
	// Request host is an resolved IP. TLS/SSL handshakes may require
	// the original server name, keep it to initate the TLS client.
	OriginalHost string
}

// Creates a req object from req options
func (r *ReqOpts) Request() *http.Request {
	req, _ := http.NewRequest(r.Method, r.Url, strings.NewReader(r.Body))
	req.Header = r.Header

	// update the Host value in the Request - this is used as the host header in any subsequent request
	req.Host = r.OriginalHost

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
	ProxyAddr string

	bar     *pb.ProgressBar
	results chan *result
}

func newPb(size int) (bar *pb.ProgressBar) {
	bar = pb.New(size)
	bar.Format("Bom !")
	bar.Start()
	return
}
