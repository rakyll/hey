package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/rakyll/boom/commands"
)

var (
	flagMethod  = flag.String("m", "GET", "")
	flagHeaders = flag.String("h", "", "")
	flagD       = flag.String("d", "", "")
	flagType    = flag.String("t", "text/html", "")
	flagAuth    = flag.String("a", "", "")

	flagC = flag.Int("c", 50, "")
	flagN = flag.Int("n", 200, "")
)

var usage = `Usage: boom [options...] <url>

Options:
  -n	Number of requests to run.
  -c	Number of requests to run concurrently. Total number of requests cannot
  	be smaller than the concurency level.

  -m	HTTP method, one of GET, POST, PUT, DELETE, HEAD, OPTIONS.
  -h	Custom HTTP headers, name1:value1;name2:value2.
  -d	HTTP request body.
  -t	Content-type, defaults to "text/html".
  -a	Basic authentication, username:password.
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	flag.Parse()
	if flag.NArg() < 1 {
		usageAndExit()
	}

	n := *flagN
	c := *flagC

	if c > n {
		usageAndExit()
	}

	url := flag.Args()[0]
	method := strings.ToUpper(*flagMethod)
	req, _ := http.NewRequest(method, url, strings.NewReader(*flagD))

	// set content-type
	req.Header.Set("Content-Type", *flagType)

	// set any other additional headers
	if *flagHeaders != "" {
		headers := strings.Split(*flagHeaders, ";")
		for _, h := range headers {
			re := regexp.MustCompile("(\\w+):(\\w+)")
			matches := re.FindAllStringSubmatch(h, -1)
			if len(matches) < 1 {
				usageAndExit()
			}
			req.Header.Set(matches[0][1], matches[0][2])
		}
	}

	// set basic auth if set
	if *flagAuth != "" {
		re := regexp.MustCompile("(\\w+):(\\w+)")
		matches := re.FindAllStringSubmatch(*flagAuth, -1)
		if len(matches) < 1 {
			usageAndExit()
		}
		req.SetBasicAuth(matches[0][1], matches[0][2])
	}

	(&commands.Boom{N: n, C: c, Req: req}).Run()
}

func usageAndExit() {
	flag.Usage()
	os.Exit(1)
}
