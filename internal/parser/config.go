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

package parser

import "time"

type Config struct {
	Url string
	// method
	M           string
	HeaderSlice []string
	Body        string
	BodyFile    string
	AuthHeader  string
	UserAgent   string
	Debug       bool

	Output string

	// concurrency
	C int
	// number of request
	N int
	// rate
	Q float64
	// timeout
	T   int
	Dur time.Duration

	H2   bool
	Cpus int

	DisableCompression bool
	DisableKeepAlives  bool
	DisableRedirects   bool
	ProxyAddr          string
}
