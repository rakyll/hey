package commands

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/cheggaaa/pb"
)

func (b *Boom) Run() error {
	b.Req.Header.Add("cache-control", "no-cache")
	b.init()
	b.run()
	b.teardown()

	b.summary()
	return nil
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

func (b *Boom) summary() {
	total := b.end.Sub(b.start)
	totalSuccessful := 0
	var avgTotal int64
	var fastest, slowest time.Duration

	for {
		select {
		case r := <-b.results:
			avgTotal += r.dur.Nanoseconds()
			if fastest.Nanoseconds() == 0 || r.dur.Nanoseconds() < fastest.Nanoseconds() {
				fastest = r.dur
			}
			if r.dur.Nanoseconds() > slowest.Nanoseconds() {
				slowest = r.dur
			}
			if r.is2xx {
				totalSuccessful++
			}
		default:
			rps := float64(b.N) / total.Seconds()
			fmt.Printf("\n")
			fmt.Printf("Results:\n")
			fmt.Printf("  total:\t%v requests\n", b.N)
			fmt.Printf("  concurrency:\t%v concurrent requests\n", b.C)
			fmt.Printf("  total 2xx:\t%v requests\n", totalSuccessful)
			fmt.Printf("  total:\t%v\n", total.Nanoseconds())
			fmt.Printf("  slowest:\t%v\n", slowest.Nanoseconds())
			fmt.Printf("  fastest:\t%v\n", fastest.Nanoseconds())
			fmt.Printf("  average:\t%v\n", avgTotal/int64(b.N))
			fmt.Printf("  requests/sec:\t%v\n", rps)
			fmt.Printf("  speed index:\t%v\n", speedIndex(rps))
			return
		}
	}
}

func (b *Boom) run() {
	var wg sync.WaitGroup
	wg.Add(b.N)

	for i := 0; i < b.N; i++ {
		go func() {
			b.runOneReq()
			b.bar.Increment()
			wg.Done()
		}()
	}

	wg.Wait()
}

func (b *Boom) runOneReq() {
	s := time.Now()
	resp, err := b.Client.Do(b.Req)

	is2xx := resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300
	b.results <- &result{is2xx: is2xx, err: err, dur: time.Now().Sub(s)}
}

func (b *Boom) handleResp(resp *http.Response, err error, dur time.Duration) {
	is2xx := false
	if resp != nil {
		is2xx = resp.StatusCode >= 200 && resp.StatusCode < 300
	}
	b.results <- &result{is2xx: is2xx, err: err, dur: dur}
}

func speedIndex(rps float64) string {
	if rps > 500 {
		return ""
	} else if rps > 100 {
		return "Pretty good"
	} else if rps > 50 {
		return "Meh"
	} else {
		return "Hahahaha"
	}
}
