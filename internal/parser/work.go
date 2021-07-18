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
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"runtime"
	"strings"

	gourl "net/url"

	"github.com/rakyll/hey/requester"
)

const (
	defaultUserAgent = "hey/0.0.1"
)

func NewWork(conf *Config) (*requester.Work, error) {
	runtime.GOMAXPROCS(conf.Cpus)
	err := validate(conf)
	if err != nil {
		return nil, err
	}

	if conf.Dur > 0 {
		conf.N = math.MaxInt32
	}

	var proxyURL *gourl.URL
	if conf.ProxyAddr != "" {
		var err error
		proxyURL, err = gourl.Parse(conf.ProxyAddr)
		if err != nil {
			return nil, err
		}
	}

	method := strings.ToUpper(conf.M)

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

	header, err := newHttpHeader(conf)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, conf.Url, nil)
	if err != nil {
		return nil, err
	}

	req.ContentLength = int64(len(bodyAll))
	req.Header = header
	return &requester.Work{
		Request:            req,
		RequestBody:        bodyAll,
		N:                  conf.N,
		Debug:              conf.Debug,
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

func newHttpHeader(conf *Config) (http.Header, error) {
	header := make(http.Header)
	setHeader(header, "Content-Type", conf.ContentType)
	setHeader(header, "Accept", conf.Accept)
	setHeader(header, "Host", conf.HostHeader)
	setHeader(header, "User-Agent", conf.UserAgent)

	if conf.AuthHeader != "" {
		match, err := parseInputWithRegexp(conf.AuthHeader, authRegexp)
		if err != nil {
			return nil, err
		}
		header.Set("Authorization", "Basic "+basicAuth(match[1], match[2]))
	}

	// set any other additional repeatable headers
	for _, h := range conf.HeaderSlice {
		match, err := parseInputWithRegexp(h, headerRegexp)
		if err != nil {
			return nil, err
		}
		header.Set(match[1], match[2])
	}
	ua := header.Get("User-Agent")
	if ua == "" {
		header.Set("User-Agent", defaultUserAgent)
	}
	return header, nil
}

func validate(conf *Config) error {
	if conf.N < conf.C {
		fmt.Printf("-c is larger than -n. Setting -c to (%v) instead.\n", conf.N)
		conf.C = conf.N
	}

	if conf.N <= 0 || conf.C <= 0 {
		return errors.New("-n and -c must be greater than 1")
	}
	return nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func setHeader(h http.Header, header string, value string) {
	if value != "" {
		h.Set(header, value)
	}
}
