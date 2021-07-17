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

package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"
	"time"
)

var (
	args  = os.Args[1:]
	flags = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
)

const usage = `Usage: hey [options...] <url>

Options:
  -n  Number of requests to run. Default is 200.
  -c  Number of workers to run concurrently. Total number of requests cannot
      be smaller than the concurrency level. Default is 50.
  -q  Rate limit, in queries per second (QPS) per worker. Default is no rate limit.
  -z  Duration of application to send requests. When duration is reached,
      application stops and exits. If duration is specified, n is ignored.
      Examples: -z 10s -z 3m.
  -o  Output type. If none provided, a summary is printed.
      "csv" is the only supported alternative. Dumps the response
      metrics in comma-separated values format.

  -m  HTTP method, one of GET, POST, PUT, DELETE, HEAD, OPTIONS.
  -H  Custom HTTP header. You can specify as many as needed by repeating the flag.
      For example, -H "Accept: text/html" -H "Content-Type: application/xml" .
  -t  Timeout for each request in seconds. Default is 20, use 0 for infinite.
  -A  HTTP Accept header.
  -d  HTTP request body.
  -D  HTTP request body from file. For example, /home/user/file.txt or ./file.txt.
  -T  Content-type, defaults to "text/html".
  -U  User-Agent, defaults to version "hey/0.0.1".
  -a  Basic authentication, username:password.
  -x  HTTP Proxy address as host:port.
  -h2 Enable HTTP/2.

  -host	HTTP Host header.

  -disable-compression  Disable compression.
  -disable-keepalive    Disable keep-alive, prevents re-use of TCP
                        connections between different HTTP requests.
  -disable-redirects    Disable following of HTTP redirects
  -cpus                 Number of used cpu cores.
                        (default for current machine is %d cores)
`

// Parses the command-line arguments provided to the program.
func parseFlags() (*Config, error) {
	flags.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage, runtime.NumCPU()))
	}

	var conf Config
	flags.StringVar(&conf.m, "m", "GET", "")
	// TODO: Remove support this deprecated flag
	h := flags.String("h", "", "")
	flags.Var(&conf.headerSlice, "H", "")
	flags.StringVar(&conf.body, "d", "", "")
	flags.StringVar(&conf.bodyFile, "D", "", "")
	flags.StringVar(&conf.accept, "A", "", "")
	flags.StringVar(&conf.contentType, "T", "text/html", "")
	flags.StringVar(&conf.authHeader, "a", "", "")
	flags.StringVar(&conf.hostHeader, "host", "", "")
	flags.StringVar(&conf.userAgent, "U", "", "")
	flags.StringVar(&conf.output, "o", "", "")

	flags.IntVar(&conf.c, "c", 50, "")
	flags.IntVar(&conf.n, "n", 200, "")
	flags.Float64Var(&conf.q, "q", 0, "")
	flags.IntVar(&conf.t, "t", 20, "")
	flags.DurationVar(&conf.dur, "z", 0, "")

	flags.BoolVar(&conf.h2, "h2", false, "")
	flags.IntVar(&conf.cpus, "cpus", runtime.GOMAXPROCS(-1), "")
	flags.BoolVar(&conf.disableCompression, "disable-compression", false, "")
	flags.BoolVar(&conf.disableKeepAlives, "disable-keepalive", false, "")
	flags.BoolVar(&conf.disableRedirects, "disable-redirects", false, "")
	flags.StringVar(&conf.proxyAddr, "x", "", "")

	flags.Parse(args)

	if flags.NArg() < 1 {
		return nil, errors.New("")
	}
	conf.url = flags.Args()[0]

	if conf.dur > 0 {
		conf.n = math.MaxInt32
		if conf.c <= 0 {
			return nil, errors.New("-c cannot be smaller than 1")
		}
	} else {
		if conf.n <= 0 || conf.c <= 0 {
			return nil, errors.New("-n and -c cannot be smaller than 1")
		}
		if conf.n < conf.c {
			return nil, errors.New("-n cannot be less than -c")
		}
	}

	// TODO: Remove support this deprecated flag
	if *h != "" {
		return nil, errors.New("flag '-h' is deprecated, please use '-H' instead")
	}

	return &conf, nil
}

type Config struct {
	url         string
	m           string
	headerSlice headerSlice
	body        string
	bodyFile    string
	accept      string
	contentType string
	authHeader  string
	hostHeader  string
	userAgent   string

	output string

	c   int
	n   int
	q   float64
	t   int
	dur time.Duration

	h2   bool
	cpus int

	disableCompression bool
	disableKeepAlives  bool
	disableRedirects   bool
	proxyAddr          string
}

type headerSlice []string

func (h *headerSlice) String() string {
	return fmt.Sprintf("%s", *h)
}

func (h *headerSlice) Set(value string) error {
	*h = append(*h, value)
	return nil
}
