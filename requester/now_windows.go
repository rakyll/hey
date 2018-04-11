// Copyright 2018 Google Inc. All Rights Reserved.
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
	"syscall"
	"time"
	"unsafe"
)

// now returns time.Duration using queryPerformanceCounter
func now() time.Duration {
	var now int64
	syscall.Syscall(queryPerformanceCounterProc.Addr(), 1, uintptr(unsafe.Pointer(&now)), 0, 0)
	return time.Duration(now) * time.Second / (time.Duration(qpcFrequency) * time.Nanosecond)
}

// precision timing
var (
	modkernel32                   = syscall.NewLazyDLL("kernel32.dll")
	queryPerformanceFrequencyProc = modkernel32.NewProc("QueryPerformanceFrequency")
	queryPerformanceCounterProc   = modkernel32.NewProc("QueryPerformanceCounter")

	qpcFrequency = queryPerformanceFrequency()
)

// queryPerformanceFrequency returns frequency in ticks per second
func queryPerformanceFrequency() int64 {
	var freq int64
	r1, _, _ := syscall.Syscall(queryPerformanceFrequencyProc.Addr(), 1, uintptr(unsafe.Pointer(&freq)), 0, 0)
	if r1 == 0 {
		panic("call failed")
	}
	return freq
}
