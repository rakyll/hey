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

// Package requester provides commands to run load tests and display results.
package requester

import (
	"os"
	"sync"
)

type InnerWriter struct {
	mutex    sync.Mutex
	FilePath string
	handler  *os.File
}

func (w *InnerWriter) Init() error {
	f, err := os.OpenFile(w.FilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	w.handler = f
	return nil
}

func (w *InnerWriter) Write(p []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	n, err = w.handler.Write(p)
	return n, err
}

func (w *InnerWriter) Close() error {
	if w.handler != nil {
		err := w.handler.Close()
		return err
	}
	return nil
}

func NewInnerWriter(fpath string) *InnerWriter {
	return &InnerWriter{
		FilePath: fpath,
	}
}
