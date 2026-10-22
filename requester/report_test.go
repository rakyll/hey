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

package requester

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestSnapshotMinMaxFields(t *testing.T) {

	r := &report{
		lats:      []float64{0.5, 0.2, 0.9, 0.3, 0.7},
		connLats:  []float64{0.030, 0.010, 0.050, 0.020, 0.040},
		dnsLats:   []float64{0.006, 0.002, 0.010, 0.004, 0.008},
		reqLats:   []float64{0.0003, 0.0001, 0.0005, 0.0002, 0.0004},
		delayLats: []float64{0.300, 0.100, 0.500, 0.200, 0.400},
		resLats:   []float64{0.060, 0.020, 0.100, 0.040, 0.080},
	}

	s := r.snapshot()

	tests := []struct {
		name string
		got  float64
		want float64
	}{
		{"ConnMin", s.ConnMin, 0.010},
		{"ConnMax", s.ConnMax, 0.050},
		{"DnsMin", s.DnsMin, 0.002},
		{"DnsMax", s.DnsMax, 0.010},
		{"ReqMin", s.ReqMin, 0.0001},
		{"ReqMax", s.ReqMax, 0.0005},
		{"DelayMin", s.DelayMin, 0.100},
		{"DelayMax", s.DelayMax, 0.500},
		{"ResMin", s.ResMin, 0.020},
		{"ResMax", s.ResMax, 0.100},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

func TestDetailsRenderedFastestBeforeSlowest(t *testing.T) {
	r := &report{
		lats:      []float64{0.5, 0.2, 0.9, 0.3, 0.7},
		connLats:  []float64{0.030, 0.010, 0.050, 0.020, 0.040},
		dnsLats:   []float64{0.006, 0.002, 0.010, 0.004, 0.008},
		reqLats:   []float64{0.0003, 0.0001, 0.0005, 0.0002, 0.0004},
		delayLats: []float64{0.300, 0.100, 0.500, 0.200, 0.400},
		resLats:   []float64{0.060, 0.020, 0.100, 0.040, 0.080},
	}
	s := r.snapshot()

	var buf bytes.Buffer
	if err := newTemplate("").Execute(&buf, s); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		prefix   string
		min, max float64
	}{
		{"DNS+dialup:", s.ConnMin, s.ConnMax},
		{"DNS-lookup:", s.DnsMin, s.DnsMax},
		{"req write:", s.ReqMin, s.ReqMax},
		{"resp wait:", s.DelayMin, s.DelayMax},
		{"resp read:", s.ResMin, s.ResMax},
	}

	for _, tt := range tests {
		var line string
		for _, l := range strings.Split(buf.String(), "\n") {
			if strings.Contains(l, tt.prefix) {
				line = l
				break
			}
		}
		if line == "" {
			t.Errorf("%s line not found in output", tt.prefix)
			continue
		}
		minStr := fmt.Sprintf("%4.4f", tt.min)
		maxStr := fmt.Sprintf("%4.4f", tt.max)
		minIdx := strings.Index(line, minStr)
		maxIdx := strings.Index(line, maxStr)
		if minIdx < 0 || maxIdx < 0 || minIdx > maxIdx {
			t.Errorf("%s: want %s before %s, got %q", tt.prefix, minStr, maxStr, line)
		}
	}
}
