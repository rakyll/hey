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
	"encoding/base64"
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	gourl "net/url"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/amit-handda/hey/requester"
	"github.com/spf13/cobra"
)

var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "benchmark get",
	Run: getFunc,
}
var (
	gKeySize      int
	gKeySpaceSize int
)

func init() {
	RootCmd.AddCommand( GetCmd)
	GetCmd.Flags().IntVar(&gKeySize, "key-size", 8, "Key size of get request")
	GetCmd.Flags().IntVar(&gKeySpaceSize, "key-space-size", 10000000, "Maximum possible keys")
}

func getFunc(cmd *cobra.Command, args []string) {
	var hs headerSlice
	flag.Var(&hs, "H", "")

	flag.Parse()
	if flag.NArg() < 1 {
		usageAndExit("")
	}

	runtime.GOMAXPROCS(cpus)
	num := n
	conc := c
	q := q
	dur := z

	if dur > 0 {
		num = math.MaxInt32
		if conc <= 0 {
			usageAndExit("-c cannot be smaller than 1.")
		}
	} else {
		if num <= 0 || conc <= 0 {
			usageAndExit("-n and -c cannot be smaller than 1.")
		}

		if num < conc {
			usageAndExit("-n cannot be less than -c.")
		}
	}

	method := "GET"

	// set content-type
	header := make(http.Header)
	header.Set("Content-Type", contentType)
	// set any other additional headers
	if headers != "" {
		usageAndExit("Flag '-h' is deprecated, please use '-H' instead.")
	}
	// set any other additional repeatable headers
	for _, h := range hs {
		match, err := parseInputWithRegexp(h, headerRegexp)
		if err != nil {
			usageAndExit(err.Error())
		}
		header.Set(match[1], match[2])
	}

	if accept != "" {
		header.Set("Accept", accept)
	}

	// set basic auth if set
	var username, password string
	if authHeader != "" {
		match, err := parseInputWithRegexp(authHeader, authRegexp)
		if err != nil {
			usageAndExit(err.Error())
		}
		username, password = match[1], match[2]
	}

	var bodyAll []byte
	if body != "" {
		bodyAll = []byte(body)
	}
	if bodyFile != "" {
		slurp, err := ioutil.ReadFile(bodyFile)
		if err != nil {
			errAndExit(err.Error())
		}
		bodyAll = slurp
	}

	var proxyURL *gourl.URL
	if proxyAddr != "" {
		var err error
		proxyURL, err = gourl.Parse(proxyAddr)
		if err != nil {
			usageAndExit(err.Error())
		}
	}

	req, err := http.NewRequest(method, args[0], nil)
	if err != nil {
		usageAndExit(err.Error())
	}
	bodyAll = bodyAll
	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}

	// set host header if set
	if hostHeader != "" {
		req.Host = hostHeader
	}

	ua := req.UserAgent()
	if ua == "" {
		ua = heyUA
	} else {
		ua += " " + heyUA
	}
	header.Set("User-Agent", ua)
	req.Header = header

	w := &requester.Work{
		Request:            req,
		RequestURL:			func() string {
			k := make([]byte, gKeySize)
			binary.PutVarint(k, int64(rand.Intn(gKeySpaceSize)))
			return fmt.Sprintf( "%s/v1/kv/%s", args[0], base64.StdEncoding.EncodeToString(k))
		},
		RequestBody:        func() []byte {
			return nil
		},
		N:                  num,
		C:                  conc,
		QPS:                q,
		Timeout:            t,
		DisableCompression: disableCompression,
		DisableKeepAlives:  disableKeepAlives,
		DisableRedirects:   disableRedirects,
		H2:                 h2,
		ProxyAddr:          proxyURL,
		Output:             output,
	}
	w.Init()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		w.Stop()
	}()
	if dur > 0 {
		go func() {
			time.Sleep(dur)
			w.Stop()
		}()
	}
	w.Run()
}

