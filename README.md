# hey

[![Build Status](https://travis-ci.org/rakyll/hey.png?branch=master)](https://travis-ci.org/rakyll/hey)

hey is a tiny program that sends some load to a web application.
It's similar to Apache Bench ([ab](http://httpd.apache.org/docs/2.2/programs/ab.html)),
but with better availability across different platforms and a less troubling
installation experience if you have [Go](http://golang.org/) installed.

hey was originally called boom and was influnced from Tarek Ziade's
tool at [tarekziade/boom](https://github.com/tarekziade/boom).

Using the same name was a mistake as it resulted in cases
where binary name conflicts created confusion.
To preserve the name for its original owner, we renamed this project to hey.

## Installation

    go get github.com/rakyll/hey

## Usage

hey runs provided number of requests in the provided concurrency level and prints stats.

It also supports HTTP2 endpoints.

```
Usage: hey [options...] <url>

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

  -h2  Make HTTP/2 requests.

  -disable-compression  Disable compression.
  -disable-keepalive    Disable keep-alive, prevents re-use of TCP
                        connections between different HTTP requests.
  -cpus                 Number of used cpu cores.
                        (default for current machine is 8 cores)
  -host                 HTTP Host header.
  ```
