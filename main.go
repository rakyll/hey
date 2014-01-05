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
	flagMethod  = flag.String("m", "GET", "http method: {GET,POST,DELETE,PUT,HEAD,OPTIONS}")
	flagHeaders = flag.String("h", "", "custom headers name1:value1,name2:value2")
	flagD       = flag.String("d", "", "request body")
	flagAuth    = flag.String("auth", "", "basic authentication user:password")

	flagC = flag.Int("c", 50, "concurrency")
	flagN = flag.Int("n", 200, "number of requests to make")
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "no url provided\n")
		os.Exit(1)
	}

	url := flag.Args()[0]
	method := strings.ToUpper(*flagMethod)
	n := *flagN
	c := *flagC

	if c > n {
		fmt.Fprintf(os.Stderr, "total number of requests to make cannot be smaller than concurrency level\n")
		os.Exit(1)
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}

	(&commands.Boom{N: n, C: c, Req: req}).Run()
}
