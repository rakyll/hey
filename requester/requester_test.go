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
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestN(t *testing.T) {
	var count int64
	handler := func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, int64(1))
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	w := &Work{
		Request: req,
		N:       20,
		C:       2,
	}
	w.Run()
	if count != 20 {
		t.Errorf("Expected to send 20 requests, found %v", count)
	}
}

func TestQps(t *testing.T) {
	var wg sync.WaitGroup
	var count int64
	handler := func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, int64(1))
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	w := &Work{
		Request: req,
		N:       20,
		C:       2,
		QPS:     1,
	}
	wg.Add(1)
	time.AfterFunc(time.Second, func() {
		if count > 2 {
			t.Errorf("Expected to work at most 2 times, found %v", count)
		}
		wg.Done()
	})
	go w.Run()
	wg.Wait()
}

func TestRequest(t *testing.T) {
	var uri, contentType, some, auth string
	handler := func(w http.ResponseWriter, r *http.Request) {
		uri = r.RequestURI
		contentType = r.Header.Get("Content-type")
		some = r.Header.Get("X-some")
		auth = r.Header.Get("Authorization")
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	header := make(http.Header)
	header.Add("Content-type", "text/html")
	header.Add("X-some", "value")
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header = header
	req.SetBasicAuth("username", "password")
	w := &Work{
		Request: req,
		N:       1,
		C:       1,
	}
	w.Run()
	if uri != "/" {
		t.Errorf("Uri is expected to be /, %v is found", uri)
	}
	if contentType != "text/html" {
		t.Errorf("Content type is expected to be text/html, %v is found", contentType)
	}
	if some != "value" {
		t.Errorf("X-some header is expected to be value, %v is found", some)
	}
	if auth != "Basic dXNlcm5hbWU6cGFzc3dvcmQ=" {
		t.Errorf("Basic authorization is not properly set")
	}
}

func TestBody(t *testing.T) {
	var count int64
	handler := func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		if string(body) == "Body" {
			atomic.AddInt64(&count, 1)
		}
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	req, _ := http.NewRequest("POST", server.URL, bytes.NewBuffer([]byte("Body")))
	w := &Work{
		Request:     req,
		RequestBody: []byte("Body"),
		N:           10,
		C:           1,
	}
	w.Run()
	if count != 10 {
		t.Errorf("Expected to work 10 times, found %v", count)
	}
}

func TestCert(t *testing.T) {
	var count int64
	const clientCertFile = "unittestClient.crt"
	const clientKeyFile = "unittestClient.key"
	const serverCertFile = "unittestServer.crt"
	const serverKeyFile = "unittestServer.key"

	// Set up and run server
	go func() {
		// Route
		handler := func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt64(&count, int64(1))
		}
		mux := http.NewServeMux()
		mux.HandleFunc("/", handler)

		// Cert validation
		cert, _ := ioutil.ReadFile(clientCertFile)
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(cert)
		tlsConfig := &tls.Config{
			ClientCAs:  certPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		}
		tlsConfig.BuildNameToCertificate()

		server := &http.Server{
			Addr:      ":7788",
			Handler:   mux,
			TLSConfig: tlsConfig,
		}

		// Note client does not need to validate the server cert
		// because `Work` has `InsecureSkipVerify: true`
		// Here we specify server cert just to make it a HTTPS server
		err := server.ListenAndServeTLS(serverCertFile, serverKeyFile)
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Failed to start HTTPS server: %v", err)
		}
	}()

	// Have this just to ensure the server is up and running
	time.Sleep(100)

	// Set up and run clients
	const numOfRun int64 = 20
	cert, _ := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	req, _ := http.NewRequest("GET", "https://localhost:7788/", nil)
	w := &Work{
		Request: req,
		N:       int(numOfRun),
		C:       2,
		Cert:    &cert,
	}
	w.Run()

	// Assert on number of requests handled by the server
	// Note the test should have failed before here with `tls: bad certificate`
	// if `Worker` does not handle `Cert` properly
	if count != numOfRun {
		t.Errorf("Expected to send %v requests, found %v", numOfRun, count)
	}
}
