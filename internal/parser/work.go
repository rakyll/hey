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
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"runtime"
	"strings"

	gourl "net/url"

	"github.com/angkeith/hey/internal/config"
	"github.com/angkeith/hey/requester"
)

const (
	defaultUserAgent = "hey/0.0.1"
)

func NewWork(conf *config.Config) (*requester.Work, error) {
	runtime.GOMAXPROCS(conf.Cpus)
	err := validate(conf)
	if err != nil {
		return nil, err
	}

	conf = sensibleDefaultOverrides(conf)

	req, err := http.NewRequest(conf.M, conf.Url, nil)
	if err != nil {
		return nil, err
	}

	var proxyURL *gourl.URL
	if conf.ProxyAddr != "" {
		var err error
		proxyURL, err = gourl.Parse(conf.ProxyAddr)
		if err != nil {
			return nil, err
		}
	}

	var body []byte
	if conf.Data != "" {
		if strings.HasPrefix(conf.Data, "@") {
			slurp, err := ioutil.ReadFile(conf.Data[1:])
			if err != nil {
				return nil, err
			}
			body = slurp
		} else {
			body = []byte(conf.Data)
		}
	}

	header, err := newHttpHeader(conf)
	if err != nil {
		return nil, err
	}

	req.ContentLength = int64(len(body))
	req.Header = header
	if h := header.Get("Host"); h != "" {
		req.Host = h
	}
	return &requester.Work{
		Request:            req,
		RequestBody:        body,
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

func newHttpHeader(conf *config.Config) (http.Header, error) {
	header := make(http.Header)
	if conf.UserAgent != "" {
		header.Set("User-Agent", conf.UserAgent)
	}

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

func validate(c *config.Config) error {
	if c.N <= 0 || c.C <= 0 {
		return errors.New("-n and -c must be greater than 1")
	}
	return nil
}

// Overrides some of the default value
func sensibleDefaultOverrides(c *config.Config) *config.Config {
	c.M = strings.ToUpper(c.M)

	if c.N < c.C {
		fmt.Printf("-c is larger than -n. Setting -c to (%v) instead.\n", c.N)
		c.C = c.N
	}

	if c.Dur > 0 {
		c.N = math.MaxInt32
	}

	if c.Data != "" && c.M == "" {
		c.M = "POST"
		c.HeaderSlice = append([]string{"Content-Type: application/x-www-form-urlencoded"}, c.HeaderSlice...)
	}

	if c.Debug {
		c.N = 1
		c.C = 1
	}

	return c
}
