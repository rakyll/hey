![hey](http://i.imgur.com/szzD9q0.png)

hey is a tiny program that sends some load to a web application.

hey was originally called boom and was influenced from Tarek Ziade's
tool at [tarekziade/boom](https://github.com/tarekziade/boom). Using the same name was a mistake as it resulted in cases
where binary name conflicts created confusion.
To preserve the name for its original owner, we renamed this project to hey.

## Installation

  - Linux (amd64): https://storage.googleapis.com/hey-releases/hey_linux_amd64
  - macOS (amd64): https://storage.googleapis.com/hey-releases/hey_darwin_amd64
  - Windows (amd64): https://storage.googleapis.com/hey-releases/hey_windows_amd64

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
  -x  HTTP Proxy address as proto://host:port. For example, http://127.0.0.1:8080
  -h2 Enable HTTP/2.

  -host	HTTP Host header.

  -disable-compression  Disable compression.
  -disable-keepalive    Disable keep-alive, prevents re-use of TCP
                        connections between different HTTP requests.
  -disable-redirects    Disable following of HTTP redirects
  -cpus                 Number of used cpu cores.
                        (default for current machine is 8 cores)
```

## Examples

Make requests with default settings:
```
hey https://google.com
```

Make 1000 requests with 100 concurrent workers:
```
hey -n 1000 -c 100 https://google.com
```

Run load test for 30 seconds:
```
hey -z 30s https://google.com
```

Make POST request with custom body:
```
hey \
    -m POST \
    -d "param1=value1&param2=value2" \
    https://google.com
```

Add custom headers:
```
hey \
    -H "Accept: application/json" \
    -H "Authorization: Bearer token" \
    https://google.com
```

Test with HTTP/2:
```
hey -h2 https://google.com
```

Rate limit to 10 queries per second per worker:
```
hey -q 10 -c 5 -z 30s https://google.com
```
