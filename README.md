# boom

`boom` is a tiny program that sends some load to a web application. It's similar to Apache Benchmark, but with better availability across different platforms and a less troubling installation experience.

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
  
	$ boom -n 300 -c 100 http://google.com
	300 / 300 [=================================================] 100.00 %
	
	Summary:
	  total:        63.005635867 secs
	  slowest:      39.885779975 secs
	  fastest:      1.119809052 secs
	  average:      7.823576417270001e+18 secs
	  requests/sec: 4.761478808550979
	  speed index:  Hahahaha
	
	Status code distribution:
	  [200]	233 responses
	  [503]	67 responses
	  
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
limitations under the License.