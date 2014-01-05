package commands

import (
	"net/http"
	"time"

	"github.com/cheggaaa/pb"
)

type result struct {
	err        error
	statusCode int
	duration   time.Duration
}

type Boom struct {
	Req    *http.Request
	N      int
	C      int
	Client *http.Client

	start   time.Time
	end     time.Time
	results chan *result
	bar     *pb.ProgressBar
}
