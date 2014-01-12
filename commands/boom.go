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
	"net/http"
	"time"

	"github.com/cheggaaa/pb"
)

type result struct {
	err        error
	statusCode int
	duration   time.Duration
}

type Boom struct {
	// Request to make.
	Req *http.Request
	// Total number of requests to make.
	N int
	// Concurrency level, the number of concurrent workers to run.
	C int
	// Timeout in seconds.
	Timeout int
	// Rate limit.
	Qps int
	// HTTP client to make the requests.
	Client *http.Client
	// Option to allow insecure TLS/SSL certificates.
	AllowInsecure bool

	results chan *result
	jobs    chan bool
	bar     *pb.ProgressBar
	rpt     *report
}

func newPb(size int) (bar *pb.ProgressBar) {
	bar = pb.New(size)
	bar.Current = barChar
	bar.BarStart = ""
	bar.BarEnd = ""
	return
}
