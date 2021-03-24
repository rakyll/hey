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
	"testing"
)

var testcases = []struct {
	name string
	in   []Bucket
	out  string
}{
	{
		"Single bucket",
		[]Bucket {
			{ Mark: 0.254, Count: 55 },
		},
		"  0.254 [55]|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■\n",
	},
	{
		"Mark precision formatting",
		[]Bucket {
			{ Mark: 0.254,  Count: 55 },
			{ Mark: 1.05,   Count: 10 },
			{ Mark: 5.2,    Count: 2 },
			{ Mark: 8.5244, Count: 6 },
		},
		"  0.254 [55]|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■\n" +
		"  1.050 [10]|■■■■■■■\n" +
		"  5.200  [2]|■\n" +
		"  8.524  [6]|■■■■\n",
	},
	{
		"Mark larger than 10",
		[]Bucket {
			{ Mark: 0.254,  Count: 55 },
			{ Mark: 1.05,   Count: 10 },
			{ Mark: 8.5244, Count: 6 },
			{ Mark: 10.67,  Count: 50 },
		},
		"   0.254 [55]|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■\n" +
		"   1.050 [10]|■■■■■■■\n" +
		"   8.524  [6]|■■■■\n" +
		"  10.670 [50]|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■\n",
	},
	{
		"Mark larger than 100",
		[]Bucket {
			{ Mark: 0.254,   Count: 55 },
			{ Mark: 1.05,    Count: 10 },
			{ Mark: 8.5244,  Count: 6 },
			{ Mark: 10.67,   Count: 50 },
			{ Mark: 120.874, Count: 35 },
		},
		"    0.254 [55]|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■\n" +
		"    1.050 [10]|■■■■■■■\n" +
		"    8.524  [6]|■■■■\n" +
		"   10.670 [50]|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■\n" +
		"  120.874 [35]|■■■■■■■■■■■■■■■■■■■■■■■■■\n",
	},
	{
		"Count larger than 100,000",
		[]Bucket {
			{ Mark: 0.009, Count: 1 },
			{ Mark: 1.795, Count: 502469 },
			{ Mark: 3.580, Count: 6998 },
			{ Mark: 5.366, Count: 355 },
			{ Mark: 7.152, Count: 13 },
			{ Mark: 8.938, Count: 69 },
			{ Mark: 10.724, Count: 4 },
			{ Mark: 12.510, Count: 3 },
			{ Mark: 14.296, Count: 30 },
			{ Mark: 16.082, Count: 172 },
			{ Mark: 17.868, Count: 118 },
		},
		"   0.009      [1]|\n" +
		"   1.795 [502469]|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■\n" +
		"   3.580   [6998]|■\n" +
		"   5.366    [355]|\n" +
		"   7.152     [13]|\n" +
		"   8.938     [69]|\n" +
		"  10.724      [4]|\n" +
		"  12.510      [3]|\n" +
		"  14.296     [30]|\n" +
		"  16.082    [172]|\n" +
		"  17.868    [118]|\n",
	},
}

func TestHistogram(t *testing.T) {
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := histogram(tc.in)
			if s != tc.out {
				t.Errorf("got %q, want %q", s, tc.out)
			}
		})
	}
}

