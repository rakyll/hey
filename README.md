# boom

[![Build Status](https://travis-ci.org/rakyll/boom.png?branch=master)](https://travis-ci.org/rakyll/boom)

`boom` is a tiny program that sends some load to a web application. It's similar to Apache Bench ([ab](http://httpd.apache.org/docs/2.2/programs/ab.html)), but with better availability across different platforms and a less troubling installation experience.

`boom` is originally written by Tarek Ziade in Python and is available on [tarekziade/boom](https://github.com/tarekziade/boom). But, due to its dependency requirements and my personal annoyance of maintaining concurrent programs in Python, I decided to rewrite it in Go.

## Installation

Simple as it takes to type the following command:

    go get github.com/rakyll/boom

## Usage

boom supports custom headers, request body and basic authentication. It runs provided number of requests in the provided concurrency level, and prints stats.
    
    Usage: boom [options...] <url>
	
	Options:
	  -n	Number of requests to run.
	  -c	Number of requests to run concurrently. Total number of requests cannot
	  		be smaller than the concurency level.
	
	  -m	HTTP method, one of GET, POST, PUT, DELETE, HEAD, OPTIONS.
	  -h	Custom HTTP headers, name1:value1;name2:value2.
	  -d	HTTP request body.
	  -t	Content-type, defaults to "text/html".
	  -a	Basic authentication, username:password.
	  

This is what happens when you run boom:
  
	$ boom  http://google.com
	200 / 200 [====================] 100.00 %

	Summary:
	  Total:        1.2723 secs.
	  Slowest:      0.3447 secs.
	  Fastest:      0.2359 secs.
	  Average:      0.2714 secs.
	  Requests/sec: 157.1997
	  Speed index:  Pretty good

	Latency distribution:
	  10% in 0.2412 secs.
	  25% in 0.2551 secs.
	  50% in 0.2681 secs.
	  75% in 0.2800 secs.
	  90% in 0.3063 secs.
	  95% in 0.3078 secs.
	  99% in 0.3158 secs.

	Status code distribution:
	  [200]	197 responses
	  [503]	3 responses

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

