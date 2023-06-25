package requester

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
)

type testVal struct {
	Int  int     `json:"int"`
	Flt  float64 `json:"flt"`
	Str  string  `json:"str"`
	Int1 int     `json:"int1"`
	Flt1 float64 `json:"flt1"`
	Str1 string  `json:"str1"`
}

func TestRequestGenerator(t *testing.T) {
	t.Run("static", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "http://localhost", nil)
		body := []byte(`{"static":"content"}`)
		if RequestGenerator(req, body) != nil {
			t.Errorf("request generator must give nil for static body")
		}
	})

	t.Run("url path/param", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://localhost/{s:1:1}/{s1:2:2}?page={i:1:9}&per_page={i1:10:99}&filter={f:1:9}", nil)
		uri := strings.Replace(RequestGenerator(req, []byte{})().URL.String(), "%20", " ", -1)
		if match, _ := regexp.MatchString(`http://localhost/./..\?page=\d&per_page=\d\d&filter=\d\.\d\d$`, uri); !match {
			t.Errorf("dynamic uri path/param must be interpolated: %s", uri)
		}
	})

	t.Run("int", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "http://localhost", nil)
		fn := RequestGenerator(req, []byte(`{"int":{i:1:100}}`))
		if fn == nil {
			t.Errorf("request generator must give generator func for dynamic body with placeholder")
			return
		}
		newBody, _ := io.ReadAll(fn().Body)
		val := testVal{}
		json.Unmarshal(newBody, &val)
		if val.Int < 1 || val.Int > 100 {
			t.Errorf("i:1:100 must give integer between 1 and 100")
		}
	})

	t.Run("float", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "http://localhost", nil)
		newBody, _ := io.ReadAll(RequestGenerator(req, []byte(`{"flt":{f:10:20}}`))().Body)
		val := testVal{}
		json.Unmarshal(newBody, &val)
		if val.Flt < 10 || val.Flt > 20 {
			t.Errorf("f:10:20 must give float between 10 and 20")
		}
	})

	t.Run("string", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "http://localhost", nil)
		newBody, _ := io.ReadAll(RequestGenerator(req, []byte(`{"str":"{s:5:10}"}`))().Body)
		val := testVal{}
		json.Unmarshal(newBody, &val)
		if len(val.Str) < 5 || len(val.Str) > 10 {
			t.Errorf("s:5:10 must give string with 5 to 10 chars")
		}
	})

	t.Run("mix", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "http://localhost", nil)
		body := []byte(`{"int":{i:1:100},"flt":{f:10:20},"str":"{s:5:10}"}`)
		newBody, _ := io.ReadAll(RequestGenerator(req, body)().Body)
		val := testVal{}
		json.Unmarshal(newBody, &val)
		if val.Int < 1 || val.Int > 100 {
			t.Errorf("i:1:100 must give integer between 1 and 100")
		}
		if val.Flt < 10 || val.Flt > 20 {
			t.Errorf("f:10:20 must give float between 10 and 20")
		}
		if len(val.Str) < 5 || len(val.Str) > 10 {
			t.Errorf("s:5:10 must give string with 5 to 10 chars")
		}
	})

	t.Run("repeat", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "http://localhost", nil)
		body := []byte(`{"int":{i:1:100},"int1": {i:100:1000}, "flt":{f:20:50},"flt1":{f}, "str":"{s:25:100}","str1":"{s}"}`)
		newBody, _ := io.ReadAll(RequestGenerator(req, body)().Body)
		val := testVal{}
		json.Unmarshal(newBody, &val)
		if val.Int != val.Int1 {
			t.Errorf("same placeholder name i must give same integer")
		}
		if val.Flt != val.Flt1 {
			t.Errorf("same placeholder name f must give same float")
		}
		if val.Str != val.Str1 {
			t.Errorf("same placeholder name s must give same string")
		}
	})

	t.Run("n-repeat", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "http://localhost", nil)
		body := []byte(`{"int":{i2:10:500},"int1": {i2}, "flt":{f3:100:2000},"flt1":{f3}, "str":"{s4:50:100}","str1":"{s4}"}`)
		newBody, _ := io.ReadAll(RequestGenerator(req, body)().Body)
		val := testVal{}
		json.Unmarshal(newBody, &val)
		if val.Int != val.Int1 {
			t.Errorf("same placeholder name i2 must give same integer")
		}
		if val.Flt != val.Flt1 {
			t.Errorf("same placeholder name f3 must give same float")
		}
		if val.Str != val.Str1 {
			t.Errorf("same placeholder name s4 must give same string")
		}
	})

	t.Run("no-min:max", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "http://localhost", nil)
		body := []byte(`{"int":{i}, "flt":{f}, "str":"{s}"}`)
		newBody, _ := io.ReadAll(RequestGenerator(req, body)().Body)
		val := testVal{}
		json.Unmarshal(newBody, &val)
		if val.Int < 1 || val.Int > 10 {
			t.Errorf("i must give integer between 1 and 10")
		}
		if val.Flt < 1 || val.Flt > 10 {
			t.Errorf("f must give float between 1 and 10")
		}
		if len(val.Str) < 1 || len(val.Str) > 10 {
			t.Errorf("s must give string with 1 to 10 chars")
		}
	})

	t.Run("url path/param request body", func(t *testing.T) {
		body := []byte(`{"int":{i}, "flt":{f}, "str":"{s}"}`)
		req, _ := http.NewRequest("POST", "http://localhost/{i9:1:9}/{f9:10:99}?page={i10:1:9}", nil)
		nreq := RequestGenerator(req, body)()
		if match, _ := regexp.MatchString(`/\d/\d\d\.\d\d\?page=\d$`, nreq.URL.String()); !match {
			t.Errorf("dynamic uri path must be interpolated: %s", nreq.URL.String())
		}
		newBody, _ := io.ReadAll(RequestGenerator(req, body)().Body)
		val := testVal{}
		json.Unmarshal(newBody, &val)
		if val.Int < 1 || val.Int > 10 {
			t.Errorf("i must give integer between 1 and 10")
		}
	})
}
