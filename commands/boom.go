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

type Report struct {
	latencies      []float64
	avgTotal       float64
	statusCodeDist map[int]int
}

type Boom struct {
	Req    *http.Request
	N      int // Number of requests
	C      int // Number of Concurrent workers
	S      int // Timeout
	Q      int // Rate limit (QPS)
	Client *http.Client

	start    time.Time
	end      time.Time
	results  chan *result
	bar      *pb.ProgressBar
	timedOut bool
	report   Report
}
