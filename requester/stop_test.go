package requester

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestStopWhileWaitingForRateLimit(t *testing.T) {
	var count int64
	first := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt64(&count, 1) == 1 {
			first <- struct{}{}
		}
	}))
	defer server.Close()
	req, _ := http.NewRequest("GET", server.URL, nil)
	work := &Work{Request: req, N: 10, C: 1, QPS: 2, Writer: io.Discard}
	work.Init()
	done := make(chan struct{})
	go func() { work.Run(); close(done) }()
	select {
	case <-first:
	case <-time.After(5 * time.Second):
		t.Fatal("first request never arrived")
	}
	// Let the first response complete and the worker wait for its next tick.
	time.Sleep(50 * time.Millisecond)
	work.Stop()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("rate-limited worker did not stop")
	}
	if got := atomic.LoadInt64(&count); got != 1 {
		t.Fatalf("sent %d requests; want only the request made before Stop", got)
	}
}
