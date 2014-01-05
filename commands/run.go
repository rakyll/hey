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
	rem := b.N
	for {
		if rem == 0 {
			break
		}

		c := b.C
		if rem < b.C {
			c = rem
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
		rem -= c
	}
}

func (b *Boom) runOneReq() {
	s := time.Now()
	resp, err := b.Client.Do(b.Req)

	code := 0
	if resp != nil {
		code = resp.StatusCode
	}
	b.results <- &result{
		statusCode: code,
		duration:   time.Now().Sub(s),
		err:        err,
	}
}
