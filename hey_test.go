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
	"crypto/tls"
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseValidHeaderFlag(t *testing.T) {
	match, err := parseInputWithRegexp("X-Something: !Y10K:;(He@poverflow?)", headerRegexp)
	if err != nil {
		t.Errorf("parseInputWithRegexp errored: %v", err)
	}
	if got, want := match[1], "X-Something"; got != want {
		t.Errorf("got %v; want %v", got, want)
	}
	if got, want := match[2], "!Y10K:;(He@poverflow?)"; got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}

func TestParseInvalidHeaderFlag(t *testing.T) {
	_, err := parseInputWithRegexp("X|oh|bad-input: badbadbad", headerRegexp)
	if err == nil {
		t.Errorf("Header parsing errored; want no errors")
	}
}

func TestParseValidAuthFlag(t *testing.T) {
	match, err := parseInputWithRegexp("_coo-kie_:!!bigmonster@1969sid", authRegexp)
	if err != nil {
		t.Errorf("A valid auth flag was not parsed correctly: %v", err)
	}
	if got, want := match[1], "_coo-kie_"; got != want {
		t.Errorf("got %v; want %v", got, want)
	}
	if got, want := match[2], "!!bigmonster@1969sid"; got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}

func TestParseInvalidAuthFlag(t *testing.T) {
	_, err := parseInputWithRegexp("X|oh|bad-input: badbadbad", authRegexp)
	if err == nil {
		t.Errorf("Header parsing errored; want no errors")
	}
}

func TestParseAuthMetaCharacters(t *testing.T) {
	_, err := parseInputWithRegexp("plus+$*{:boom", authRegexp)
	if err != nil {
		t.Errorf("Auth header with a plus sign in the user name errored: %v", err)
	}
}

func TestTranslateTLSVersions(t *testing.T) {
	tests := []struct {
		tlsVersion     string
		expectedResult uint16
		expectedError  error
	}{
		{ // TLS 1.0
			tlsVersion:     "1.0",
			expectedResult: tls.VersionTLS10,
		},
		{ // TLS 1.1
			tlsVersion:     "1.1",
			expectedResult: tls.VersionTLS11,
		},
		{ // TLS 1.2
			tlsVersion:     "1.2",
			expectedResult: tls.VersionTLS12,
		},
		{ // TLS 1.3
			tlsVersion:     "1.3",
			expectedResult: tls.VersionTLS13,
		},
		{ // no version specified
			tlsVersion:     "",
			expectedResult: 0,
		},
		{ // invalid version
			tlsVersion:     "1.4",
			expectedResult: math.MaxUint16,
			expectedError:  errors.New("could not parse TLS version: 1.4"),
		},
	}

	for _, test := range tests {
		actualResult, actualError := translateTLSVersion(test.tlsVersion)
		assert.Equal(t, test.expectedResult, actualResult)
		assert.Equal(t, test.expectedError, actualError)
	}
}

func TestValidateTLSVersions(t *testing.T) {
	tests := []struct {
		minTLSVersion uint16
		maxTLSVersion uint16
		expectedError error
	}{
		{ // neither version specified
			minTLSVersion: 0,
			maxTLSVersion: 0,
		},
		{ // only min specified
			minTLSVersion: tls.VersionTLS11,
			maxTLSVersion: 0,
		},
		{ // only max specified
			minTLSVersion: 0,
			maxTLSVersion: tls.VersionTLS12,
		},
		{ // both specified
			minTLSVersion: tls.VersionTLS12,
			maxTLSVersion: tls.VersionTLS13,
		},
		{ // both specified and equal
			minTLSVersion: tls.VersionTLS13,
			maxTLSVersion: tls.VersionTLS13,
		},
		{ // invalid choices
			minTLSVersion: tls.VersionTLS12,
			maxTLSVersion: tls.VersionTLS10,
			expectedError: errors.New("min TLS version cannot be greater than max TLS version"),
		},
	}

	for _, test := range tests {
		actualError := validateTLSVersions(test.minTLSVersion, test.maxTLSVersion)
		assert.Equal(t, test.expectedError, actualError)
	}
}
