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

// Command hey is an HTTP load generator.
package cmd

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "A low-level benchmark tool for consul",
	Long: `hey-consul is a low-level benchmark tool for consul.`,
}

const (
	headerRegexp = `^([\w-]+):\s*(.+)`
	authRegexp   = `^(.+):([^\s].+)`
	heyUA        = "hey/0.0.1"
)

var (
	headers     string
	accept      string
	contentType string
	authHeader  string
	hostHeader  string

	output 		string

	c int
	n int
	q float64
	t int
	z time.Duration

	h2   bool
	cpus int

	disableCompression bool
	disableKeepAlives  bool
	disableRedirects   bool
	proxyAddr          string
)

func init() {
	RootCmd.PersistentFlags().StringVar(&headers, "h", "", "Custom HTTP header. You can specify as many as needed by repeating the flag. For example, -H 'Accept: text/html' -H 'Content-Type: application/xml' ")
	RootCmd.PersistentFlags().StringVar(&accept, "A", "", "HTTP Accept header.")
	RootCmd.PersistentFlags().StringVar(&contentType, "T", "text/html", "Content-type, defaults to 'text/html'.")
	RootCmd.PersistentFlags().StringVar(&authHeader, "a", "", "Basic authentication, username:password.")
	RootCmd.PersistentFlags().StringVar(&hostHeader, "host", "", "HTTP Host header.")

	RootCmd.PersistentFlags().StringVar(&output, "o", "", "If none provided, a summary is printed. 'csv' is the only supported alternative. Dumps the response metrics in comma-separated values format.")

	RootCmd.PersistentFlags().IntVar(&c, "c", 50, "Number of workers to run concurrently. Total number of requests cannot be smaller than the concurrency level. Default is 50.")
	RootCmd.PersistentFlags().IntVar(&n, "n", 200, "Number of requests to run. Default is 200.")
	RootCmd.PersistentFlags().Float64Var(&q, "q", 0, "Rate limit, in queries per second (QPS) per worker. Default is no rate limit.")
	RootCmd.PersistentFlags().IntVar(&t, "t", 20, "Timeout for each request in seconds. Default is 20, use 0 for infinite.")
	RootCmd.PersistentFlags().DurationVar(&z, "z", 0, "Duration of application to send requests. When duration is reached, application stops and exits. If duration is specified, n is ignored. Examples: -z 10s -z 3m.")

	RootCmd.PersistentFlags().BoolVar(&h2, "h2", false, "Enable HTTP/2." )
	RootCmd.PersistentFlags().IntVar(&cpus, "cpus", runtime.GOMAXPROCS(-1), "Number of used cpu cores. (default for current machine is %d cores)" )

	RootCmd.PersistentFlags().BoolVar(&disableCompression, "disable-compression", false, "Disable compression." )
	RootCmd.PersistentFlags().BoolVar(&disableKeepAlives, "disable-keepalive", false, "Disable keep-alive, prevents re-use of TCP connections between different HTTP requests")
	RootCmd.PersistentFlags().BoolVar(&disableRedirects, "disable-redirects", false, "Disable following of HTTP redirects")
	RootCmd.PersistentFlags().StringVar(&proxyAddr, "x", "", "HTTP Proxy address as host:port.")
}

func errAndExit(msg string) {
	fmt.Fprintf(os.Stderr, msg)
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func usageAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	//flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func parseInputWithRegexp(input, regx string) ([]string, error) {
	re := regexp.MustCompile(regx)
	matches := re.FindStringSubmatch(input)
	if len(matches) < 1 {
		return nil, fmt.Errorf("could not parse the provided input; input = %v", input)
	}
	return matches, nil
}

type headerSlice []string

func (h *headerSlice) String() string {
	return fmt.Sprintf("%s", *h)
}

func (h *headerSlice) Set(value string) error {
	*h = append(*h, value)
	return nil
}

func initRandBytes(n int, v uint64) []byte {
	k := make([]byte, n)
	for i := 0; i < n; i+=8 {
		var cv = v
		if cv == 0 {
			cv = rand.Uint64()
		}

		binary.LittleEndian.PutUint64( k[i:i+8], cv )
	}
	return k
}

func mustRandBytes(n int) []byte {
	rb := make([]byte, n)
	_, err := rand.Read(rb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate value: %v\n", err)
		os.Exit(1)
	}
	return rb
}
