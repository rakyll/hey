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

import (
	"errors"
	"io/ioutil"
	"math"
	"net/http"
	"runtime"
	"strings"

	gourl "net/url"

	"github.com/rakyll/hey/requester"
)

const (
	heyUA = "hey/0.0.1"
)

func NewWork(conf *Config) (*requester.Work, error) {
	if conf.Dur > 0 {
		conf.N = math.MaxInt32
		if conf.C <= 0 {
			return nil, errors.New("-c cannot be smaller than 1")
		}
	} else {
		if conf.N <= 0 || conf.C <= 0 {
			return nil, errors.New("-n and -c cannot be smaller than 1")
		}
		if conf.N < conf.C {
			return nil, errors.New("-n cannot be less than -c")
		}
	}

	runtime.GOMAXPROCS(conf.Cpus)

	method := strings.ToUpper(conf.M)

	// set content-type
	header := make(http.Header)
	header.Set("Content-Type", conf.ContentType)

	// set any other additional repeatable headers
	for _, h := range conf.HeaderSlice {
		match, err := parseInputWithRegexp(h, headerRegexp)
		if err != nil {
			return nil, err
		}
		header.Set(match[1], match[2])
	}

	if conf.Accept != "" {
		header.Set("Accept", conf.Accept)
	}

	// set basic auth if set
	var username, password string
	if conf.AuthHeader != "" {
		match, err := parseInputWithRegexp(conf.AuthHeader, authRegexp)
		if err != nil {
			return nil, err
		}
		username, password = match[1], match[2]
	}

	var bodyAll []byte
	if conf.Body != "" {
		bodyAll = []byte(conf.Body)
	}
	if conf.BodyFile != "" {
		slurp, err := ioutil.ReadFile(conf.BodyFile)
		if err != nil {
			return nil, err
		}
		bodyAll = slurp
	}

	var proxyURL *gourl.URL
	if conf.ProxyAddr != "" {
		var err error
		proxyURL, err = gourl.Parse(conf.ProxyAddr)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, conf.Url, nil)
	if err != nil {
		return nil, err
	}
	req.ContentLength = int64(len(bodyAll))
	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}

	// set host header if set
	if conf.HostHeader != "" {
		req.Host = conf.HostHeader
	}

	ua := header.Get("User-Agent")
	if ua == "" {
		ua = heyUA
	} else {
		ua += " " + heyUA
	}
	header.Set("User-Agent", ua)

	// set userAgent header if set
	if conf.UserAgent != "" {
		ua = conf.UserAgent + " " + heyUA
		header.Set("User-Agent", ua)
	}

	return &requester.Work{
		Request:            req,
		RequestBody:        bodyAll,
		N:                  conf.N,
		C:                  conf.C,
		QPS:                conf.Q,
		Timeout:            conf.T,
		DisableCompression: conf.DisableCompression,
		DisableKeepAlives:  conf.DisableCompression,
		DisableRedirects:   conf.DisableRedirects,
		H2:                 conf.H2,
		ProxyAddr:          proxyURL,
		Output:             conf.Output,
	}, nil
}
