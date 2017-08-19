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
