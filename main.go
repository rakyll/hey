package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/rakyll/boom/commands"
)

var (
	flagMethod  = flag.String("m", "GET", "")
	flagHeaders = flag.String("h", "", "")
	flagD       = flag.String("d", "", "")
	flagType    = flag.String("t", "text/html", "")
	// TODO: add basic auth flag

	flagC = flag.Int("c", 50, "")
	flagN = flag.Int("n", 200, "")
)

var usage = `Usage: boom [options...] <url>

Options:
  -n	Number of requests to run.
  -c 	Number of requests to run concurrently. Total number of requests cannot
  	be smaller than the concurency level.

  -m	HTTP method, one of GET, POST, PUT, DELETE, HEAD, OPTIONS.
  -h	Custom HTTP headers, name1:value1;name2:value2.
  -d	HTTP request body.
  -t	Content-type, defaults to "text/html".
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	url := flag.Args()[0]
	method := strings.ToUpper(*flagMethod)
	contentType := *flagType
	n := *flagN
	c := *flagC

	if c > n {
		flag.Usage()
		os.Exit(1)
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}

	req.Header.Set("content-type", contentType)
	(&commands.Boom{N: n, C: c, Req: req}).Run()
}
