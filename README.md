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
  
	$ boom -n 100 https://google.com
	
	100 / 100 [==============================] 100.00 %
	
	Summary:
	  Total:	222.2607 secs.
	  Slowest:	84.5743 secs.
	  Fastest:	6.7543 secs.
	  Average:	20.8187 secs.
	  Requests/sec:	0.4499
	  Speed index:	Hahahaha
	
	Response time histogram:
	  6.754 [1]	|
	  14.536 [30]	|###########################
	  22.318 [44]	|########################################
	  30.100 [14]	|############
	  37.882 [5]	|####
	  45.664 [1]	|
	  53.446 [0]	|
	  61.228 [1]	|
	  69.010 [1]	|
	  76.792 [1]	|
	  84.574 [2]	|#
	
	Latency distribution:
	  10% in 10.4375 secs.
	  25% in 13.2917 secs.
	  50% in 17.3346 secs.
	  75% in 21.9699 secs.
	  90% in 33.0562 secs.
	  95% in 39.3120 secs.
	  99% in 82.1295 secs.
	
	Status code distribution:
	  [200]	75 responses
	  [0]	25 responses

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

