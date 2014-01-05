package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

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
		fmt.Fprintf(os.Stderr, "no url provided")
	}

	url := flag.Args()[0]
	req, err := http.NewRequest(*flagMethod, url, nil)
	if err != nil {
		fmt.Errorf("error: %v", err)
		os.Exit(1)
	}

	b := &commands.Boom{N: *flagN, C: *flagC, Req: req}
	b.Run()
}
