package commands

import (
	"net/http"
	"sync"
	"time"

	"github.com/cheggaaa/pb"
)

func (b *Boom) Run() {
	b.Req.Header.Add("cache-control", "no-cache")
	b.init()
	b.run()
	b.teardown()

	b.Print()
}

func (b *Boom) init() {
	if b.Client == nil {
		b.Client = &http.Client{}
	}
	b.results = make(chan *result, b.N)
	b.bar = pb.StartNew(b.N)
	b.start = time.Now()
}

func (b *Boom) teardown() {
	b.end = time.Now()
	b.bar.Finish()
}

func (b *Boom) run() {
	remaining := b.N
	for {
		if remaining < 1 {
			break
		}

		c := b.C
		if remaining < b.C {
			c = remaining
		}

		var wg sync.WaitGroup
		wg.Add(c)
		for i := 0; i < c; i++ {
			go func() {
				b.runOneReq()
				b.bar.Increment()
				wg.Done()
			}()
		}
		wg.Wait()
		remaining = remaining - c
	}
}

func (b *Boom) runOneReq() {
	s := time.Now()
	resp, err := b.Client.Do(b.Req)

	is2xx := resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300
	b.results <- &result{is2xx: is2xx, err: err, dur: time.Now().Sub(s)}
}
