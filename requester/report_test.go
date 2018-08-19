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
	"reflect"
	"testing"
)

func TestLatencies(t *testing.T) {
	lats := make([]float64, 100)
	for i := 0; i < len(lats); i++ {
		lats[i] = float64(i + 1)
	}
	r := &report{lats: lats}
	expected := make([]LatencyDistribution, len(pctls))
	for i, pctl := range pctls {
		expected[i] = LatencyDistribution{
			Percentage: pctl,
			Latency:    float64(pctl),
		}
	}
	actual := r.latencies()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, found %v", expected, actual)
	}
}
