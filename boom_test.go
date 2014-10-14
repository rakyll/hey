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
	"testing"
)

type mockDnsResolver struct {
	Addr string
}

func (r *mockDnsResolver) Lookup(host string) ([]string, error) {
	return []string{r.Addr}, nil
}

func TestParseUrl(t *testing.T) {
	defaultDNSResolver = &mockDnsResolver{Addr: "2a00:1450:400a:806::1007"}
	u, _ := resolveUrl("http://google.com:80/path/to/resource?q=rawquery")
	if u != "http://[2a00:1450:400a:806::1007]:80/path/to/resource?q=rawquery" {
		t.Errorf("Problem during url parsing, %v is found.", u)
	}
}

func TestParseUrl_IPv4(t *testing.T) {
	defaultDNSResolver = &mockDnsResolver{Addr: "127.0.0.1"}
	u, s := resolveUrl("http://google.com")
	if s != "google.com" {
		t.Errorf("Original server name doesn't match with google.com, %v is found.", s)
	}
	if u != "http://127.0.0.1" {
		t.Errorf("URL is expected to be http://127.0.0.1, %v is found.", u)
	}
}

func TestParseUrl_IPv4AndPort(t *testing.T) {
	defaultDNSResolver = &mockDnsResolver{Addr: "127.0.0.1"}
	u, s := resolveUrl("http://google.com:80")
	if s != "google.com:80" {
		t.Errorf("Original server name doesn't match with google.com, %v is found.", s)
	}
	if u != "http://127.0.0.1:80" {
		t.Errorf("URL is expected to be http://127.0.0.1, %v is found.", u)
	}
}

func TestParseUrl_IPv6(t *testing.T) {
	defaultDNSResolver = &mockDnsResolver{Addr: "2a00:1450:400a:806::1007"}
	u, s := resolveUrl("http://google.com")
	if s != "google.com" {
		t.Errorf("Original server name doesn't match with google.com, %v is found.", s)
	}
	if u != "http://[2a00:1450:400a:806::1007]" {
		t.Errorf("URL is expected to be http://[2a00:1450:400a:806::1007], %v is found.", u)
	}
}

func TestParseUrl_IPv6AndPort(t *testing.T) {
	defaultDNSResolver = &mockDnsResolver{Addr: "2a00:1450:400a:806::1007"}
	u, s := resolveUrl("http://google.com:80")
	if s != "google.com:80" {
		t.Errorf("Original server name doesn't match with google.com, %v is found.", s)
	}
	if u != "http://[2a00:1450:400a:806::1007]:80" {
		t.Errorf("URL is expected to be http://[2a00:1450:400a:806::1007]:80, %v is found.", u)
	}
}

func TestParseValidHeaderFlag(t *testing.T) {
	match, err := parseInputWithRegexp("X-Something: !Y10K:;(He@poverflow?)", headerRegexp)
	if err != nil {
		t.Errorf("A valid header was not parsed correctly: %v", err.Error())
	}
	if match[1] != "X-Something" || match[2] != "!Y10K:;(He@poverflow?)" {
		t.Errorf("A valid header was not parsed correctly, parsed values: %v %v", match[1], match[2])
	}
}

func TestParseInvalidHeaderFlag(t *testing.T) {
	_, err := parseInputWithRegexp("X|oh|bad-input: badbadbad", headerRegexp)
	if err == nil {
		t.Errorf("An invalid header passed parsing")
	}
}

func TestParseValidAuthFlag(t *testing.T) {
	match, err := parseInputWithRegexp("_coo-kie_:!!bigmonster@1969sid", authRegexp)
	if err != nil {
		t.Errorf("A valid auth flag was not parsed correctly: %v", err.Error())
	}
	if match[1] != "_coo-kie_" || match[2] != "!!bigmonster@1969sid" {
		t.Errorf("A valid auth flag was not parsed correctly, parsed values: %v %v", match[1], match[2])
	}
}

func TestParseInvalidAuthFlag(t *testing.T) {
	_, err := parseInputWithRegexp("X|oh|bad-input: badbadbad", authRegexp)
	if err == nil {
		t.Errorf("An invalid header passed parsing")
	}
}
