![hey](http://i.imgur.com/szzD9q0.png)

[![Build Status](https://travis-ci.org/rakyll/hey.svg?branch=master)](https://travis-ci.org/rakyll/hey)

hey is a tiny program that sends some load to a web application.

hey was originally called boom and was influenced from Tarek Ziade's
tool at [tarekziade/boom](https://github.com/tarekziade/boom). Using the same name was a mistake as it resulted in cases
where binary name conflicts created confusion.
To preserve the name for its original owner, we renamed this project to hey.

## Installation

* Linux 64-bit: https://hey-release.s3.us-east-2.amazonaws.com/hey_linux_amd64
* Mac 64-bit: https://hey-release.s3.us-east-2.amazonaws.com/hey_darwin_amd64
* Windows 64-bit: https://hey-release.s3.us-east-2.amazonaws.com/hey_windows_amd64

### Package Managers

macOS:
-  [Homebrew](https://brew.sh/) users can use `brew install hey`.

## Usage

hey runs provided number of requests in the provided concurrency level and prints stats.

It also supports HTTP2 endpoints.

```
Usage: hey [options...] <url>

Options:
  -n  Number of requests to run. Default is 200.
  -c  Number of workers to run concurrently. Total number of requests cannot
      be smaller than the concurrency level. Default is 50.
  -q  Rate limit, in queries per second (QPS) per worker. Default is no rate limit.
  -z  Duration of application to send requests. When duration is reached,
      application stops and exits. If duration is specified, n is ignored.
      Examples: -z 10s -z 3m.
  -o  Output type. If none provided, a summary is printed.
      "csv" is the only supported alternative. Dumps the response
      metrics in comma-separated values format.

  -m  HTTP method, one of GET, POST, PUT, DELETE, HEAD, OPTIONS.
  -H  Custom HTTP header. You can specify as many as needed by repeating the flag.
      For example, -H "Accept: text/html" -H "Content-Type: application/xml" .
  -t  Timeout for each request in seconds. Default is 20, use 0 for infinite.
  -A  HTTP Accept header.
  -d  HTTP request body.
  -D  HTTP request body from file. For example, /home/user/file.txt or ./file.txt.
  -T  Content-type, defaults to "text/html".
  -a  Basic authentication, username:password.
  -x  HTTP Proxy address as host:port.
  -h2 Enable HTTP/2.

  -host	HTTP Host header.

  -disable-compression  Disable compression.
  -disable-keepalive    Disable keep-alive, prevents re-use of TCP
                        connections between different HTTP requests.
  -disable-redirects    Disable following of HTTP redirects
  -cpus                 Number of used cpu cores.
                        (default for current machine is 8 cores)
```

Previously known as [github.com/rakyll/boom](https://github.com/rakyll/boom).

### Dynamic request body

For method that supports body, it is possible to generate dynamic request body for both `-d` and `-D` flags.

It supports 3 basic data types: `i` for integer, `f` for float (2 decimal point) and `s` for string.
The dynamic body is generated by interpolating placeholders (`typeid:min:max`) like so:

```json
{
  "name": "{s:5:10}",  // string with 5 to 10 chars
  "age": {i:1:100},    // integer value between 1 to 100
  "score": {f:0:100},  // float value between 0 to 100
  "date": "{i1:1950:2023}-0{i2:1:9}-{i3:11:28}", // some date (yyyy-mm-dd)
}
```

#### Placeholder

A placeholder contains of 3 segments separated by colon: `typeid:min:max`.
> `min:max` is optional with defaults `min=1`, `max=10` so only `typeid` is enough.

The `typeid` segment contains data `type` (ie `i`, `f`, `s`) and optional integer suffix `(1..N)`.

Multiple occurances of same `typeid`s eg `{i1:1:10}`, `{i1}` produce same integer between 1 and 10:
> `[{i1:1:10}, {i1:0:0}, {i1}]` => `[7, 7, 7]`

Different `typeid`s eg `i1`, `i2` etc produce different values:
> `[{i1:1:10}, {i2:1:10}, {i3:1:10}]` => `[5, 1, 9]`
