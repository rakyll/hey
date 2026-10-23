// Copyright 2026 The hey Authors.
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
	"time"
)

func TestQPCDurationAfterLongUptime(t *testing.T) {
	const frequency = int64(10_000_000)
	for _, want := range []time.Duration{15 * time.Minute, time.Hour, 24*time.Hour + 500*time.Millisecond} {
		t.Run(want.String(), func(t *testing.T) {
			ticks := int64(want/time.Second)*frequency + int64(want%time.Second)*frequency/int64(time.Second)
			if got := qpcDuration(ticks, frequency); got != want {
				t.Fatalf("qpcDuration(%d, %d) = %s, want %s", ticks, frequency, got, want)
			}
		})
	}
}
