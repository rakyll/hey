package requester

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRenderer(t *testing.T) {
	var bitmap int32
	var mux sync.Mutex
	template := "{{int}} foo {{datetime}} bar {{int}}"
	start := time.Now()

	handler := func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		words := strings.Fields(string(body))
		if len(words) != 5 {
			t.Errorf("Expected to have 5 words, found %d", len(words))
		}
		if words[0] != words[4] {
			t.Errorf("Expected int to be the same, found %s %s", words[0], words[4])
		}
		if words[1]+words[3] != "foobar" {
			t.Errorf("Expected to have foo bar, found %s %s", words[1], words[3])
		}

		mux.Lock()
		defer mux.Unlock()

		renderTime, _ := time.Parse(time.RFC3339Nano, words[2])
		if renderTime.Before(start) || renderTime.After(time.Now()) {
			t.Errorf("Expected time to be recent %v, found %s", start, words[2])
		}

		i, _ := strconv.Atoi(words[0])
		if bitmap&(1<<uint(i)) != 0 {
			t.Errorf("Expected to have foo bar, found %s %s", words[1], words[3])
		}
		bitmap |= (1 << uint(i))
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	req, _ := http.NewRequest("POST", server.URL, nil)
	w := &Work{
		Request:        req,
		RequestBody:    []byte(template),
		EnableRenderer: true,
		N:              10,
		C:              1,
	}
	w.Run()

	if bitmap != (1<<uint(w.N))-1 {
		t.Errorf("Expected %d bits, found %x", w.N, bitmap)
	}
}
