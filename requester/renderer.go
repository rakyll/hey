package requester

import (
	"strconv"
	"strings"
	"time"
)

type renderer struct {
	results  chan []byte
	handlers map[string]func(id int) string
}

func newRenderer(results chan []byte) *renderer {
	return &renderer{
		results: results,
		handlers: map[string]func(int) string{
			"int": func(i int) string {
				return strconv.Itoa(i)
			},
			"datetime": func(i int) string {
				return time.Now().Format(time.RFC3339Nano)
			},
		},
	}
}

func runRenderer(r *renderer, template []byte, n int) {
	for i := 0; i < n; i++ {
		r.results <- r.Render(template, i)
	}
}

func (r *renderer) Render(template []byte, i int) []byte {
	body := string(template)
	for key, generator := range r.handlers {
		body = strings.Replace(body, "{{"+key+"}}", generator(i), -1)
	}
	return []byte(body)
}
