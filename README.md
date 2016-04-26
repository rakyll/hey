# boom

[![Build Status](https://travis-ci.org/rakyll/boom.png?branch=master)](https://travis-ci.org/rakyll/boom)

Boom is a tiny program that sends some load to a web application. It's similar to Apache Bench ([ab](http://httpd.apache.org/docs/2.2/programs/ab.html)), but with better availability across different platforms and a less troubling installation experience.

Boom is originally written by Tarek Ziade in Python and is available on [tarekziade/boom](https://github.com/tarekziade/boom). But, due to its dependency requirements and my personal annoyance of maintaining concurrent programs in Python, I decided to rewrite it in Go.

## Installation

Simple as it takes to type the following command:

    go get github.com/rakyll/boom

## Usage

Boom supports custom headers, request body and basic authentication. It runs provided number of requests in the provided concurrency level, and prints stats.
~~~
Usage: boom [options...] <url>

Options:
  -n  Number of requests to run.
  -c  Number of requests to run concurrently. Total number of requests cannot
      be smaller than the concurency level.
  -q  Rate limit, in seconds (QPS).
  -o  Output type. If none provided, a summary is printed.
      "csv" is the only supported alternative. Dumps the response
      metrics in comma-seperated values format.

  -m  HTTP method, one of GET, POST, PUT, DELETE, HEAD, OPTIONS.
  -H  Custom HTTP header. You can specify as many as needed by repeating the flag.
      for example, -H "Accept: text/html" -H "Content-Type: application/xml" .
  -t  Timeout in ms.
  -A  HTTP Accept header.
  -d  HTTP request body.
  -T  Content-type, defaults to "text/html".
  -a  Basic authentication, username:password.
  -x  HTTP Proxy address as host:port.

  -disable-compression  Disable compression.
  -disable-keepalive    Disable keep-alive, prevents re-use of TCP
                        connections between different HTTP requests.
  -cpus                 Number of used cpu cores.
                        (default for current machine is 1 cores)
~~~

This is what happens when you run Boom:

	% boom -n 1000 -c 100 https://google.com
	1000 / 1000 ∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎ 100.00 % 

	Summary:
	  Total:        21.1307 secs.
	  Slowest:      2.9959 secs.
	  Fastest:      0.9868 secs.
	  Average:      2.0827 secs.
	  Requests/sec: 47.3246
	  Speed index:  Hahahaha

	Response time histogram:
      0.987 [1]     |
      1.188 [2]     |
      1.389 [3]     |
      1.590 [18]    |∎∎
      1.790 [85]    |∎∎∎∎∎∎∎∎∎∎∎
      1.991 [244]   |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      2.192 [284]   |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      2.393 [304]   |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      2.594 [50]    |∎∎∎∎∎∎
      2.795 [5]     |
      2.996 [4]     |

	Latency distribution:
	  10% in 1.7607 secs.
	  25% in 1.9770 secs.
	  50% in 2.0961 secs.
	  75% in 2.2385 secs.
	  90% in 2.3681 secs.
	  95% in 2.4451 secs.
	  99% in 2.5393 secs.

	Status code distribution:
	  [200]	1000 responses

## License

Copyright 2014 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. ![Analytics](https://ga-beacon.appspot.com/UA-46881978-1/boom?pixel)

