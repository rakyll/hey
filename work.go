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
	"io/ioutil"
	"net/http"
	gourl "net/url"
	"runtime"
	"strings"

	"github.com/rakyll/hey/requester"
)

func newWork(conf *Config) (*requester.Work, error) {
	runtime.GOMAXPROCS(conf.cpus)

	method := strings.ToUpper(conf.m)

	// set content-type
	header := make(http.Header)
	header.Set("Content-Type", conf.contentType)

	// set any other additional repeatable headers
	for _, h := range conf.headerSlice {
		match, err := parseInputWithRegexp(h, headerRegexp)
		if err != nil {
			return nil, err
		}
		header.Set(match[1], match[2])
	}

	if conf.accept != "" {
		header.Set("Accept", conf.accept)
	}

	// set basic auth if set
	var username, password string
	if conf.authHeader != "" {
		match, err := parseInputWithRegexp(conf.authHeader, authRegexp)
		if err != nil {
			return nil, err
		}
		username, password = match[1], match[2]
	}

	var bodyAll []byte
	if conf.body != "" {
		bodyAll = []byte(conf.body)
	}
	if conf.bodyFile != "" {
		slurp, err := ioutil.ReadFile(conf.bodyFile)
		if err != nil {
			return nil, err
		}
		bodyAll = slurp
	}

	var proxyURL *gourl.URL
	if conf.proxyAddr != "" {
		var err error
		proxyURL, err = gourl.Parse(conf.proxyAddr)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, conf.url, nil)
	if err != nil {
		return nil, err
	}
	req.ContentLength = int64(len(bodyAll))
	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}

	// set host header if set
	if conf.hostHeader != "" {
		req.Host = conf.hostHeader
	}

	ua := header.Get("User-Agent")
	if ua == "" {
		ua = heyUA
	} else {
		ua += " " + heyUA
	}
	header.Set("User-Agent", ua)

	// set userAgent header if set
	if conf.userAgent != "" {
		ua = conf.userAgent + " " + heyUA
		header.Set("User-Agent", ua)
	}

	return &requester.Work{
		Request:            req,
		RequestBody:        bodyAll,
		N:                  conf.n,
		C:                  conf.c,
		QPS:                conf.q,
		Timeout:            conf.t,
		DisableCompression: conf.disableCompression,
		DisableKeepAlives:  conf.disableCompression,
		DisableRedirects:   conf.disableRedirects,
		H2:                 conf.h2,
		ProxyAddr:          proxyURL,
		Output:             conf.output,
	}, nil
}
