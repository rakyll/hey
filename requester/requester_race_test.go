package requester

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestDisableRedirectsNoRace runs many concurrent workers with DisableRedirects
// enabled. Before the fix, each worker assigned client.CheckRedirect on the
// shared *http.Client, causing a data race. Run with -race to guard against
// a regression: go test -race -run TestDisableRedirectsNoRace ./requester
func TestDisableRedirectsNoRace(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	req, err := http.NewRequest("GET", srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	w := &Work{
		Request:          req,
		N:                40,
		C:                20,
		DisableRedirects: true,
		Timeout:          10,
		Output:           "",
	}
	w.Init()
	w.Run()
}
